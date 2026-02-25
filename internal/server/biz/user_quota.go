package biz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/quotatransaction"
	"github.com/looplj/axonhub/internal/ent/userquota"
	"github.com/looplj/axonhub/internal/log"
)

type UserQuotaServiceParams struct {
	fx.In

	Ent *ent.Client
}

func NewUserQuotaService(params UserQuotaServiceParams) *UserQuotaService {
	return &UserQuotaService{
		AbstractService: &AbstractService{db: params.Ent},
	}
}

type UserQuotaService struct {
	*AbstractService
}

// UserQuotaInfo represents user quota information
type UserQuotaInfo struct {
	UserID      int     `json:"user_id"`
	Quota       int64   `json:"quota"`          // 账户余额（分）
	UsedQuota   int64   `json:"used_quota"`     // 已使用额度（分）
	Group       string  `json:"group"`          // 用户分组
	AffCode     string  `json:"aff_code"`       // 邀请码
	InviterID   *int    `json:"inviter_id"`     // 邀请人ID
	BalanceYuan float64 `json:"balance_yuan"`   // 余额（元）
}

// GetOrCreateUserQuota gets or creates user quota record
func (s *UserQuotaService) GetOrCreateUserQuota(ctx context.Context, userID int) (*ent.UserQuota, error) {
	return authz.RunWithSystemBypass(ctx, "get-or-create-quota", func(bypassCtx context.Context) (*ent.UserQuota, error) {
		client := s.entFromContext(bypassCtx)

		// Try to get existing
		quota, err := client.UserQuota.Query().
			Where(userquota.UserIDEQ(userID)).
			Only(bypassCtx)
		if err == nil {
			return quota, nil
		}
		if !ent.IsNotFound(err) {
			return nil, fmt.Errorf("failed to query user quota: %w", err)
		}

		// Create new
		affCode, err := generateAffCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate aff code: %w", err)
		}

		quota, err = client.UserQuota.Create().
			SetUserID(userID).
			SetQuota(0).
			SetUsedQuota(0).
			SetGroup("default").
			SetAffCode(affCode).
			SetNotify(true).
			SetQuotaRemindThreshold(100000). // 20元
			Save(bypassCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to create user quota: %w", err)
		}

		return quota, nil
	})
}

// GetUserQuotaInfo gets user quota info
func (s *UserQuotaService) GetUserQuotaInfo(ctx context.Context, userID int) (*UserQuotaInfo, error) {
	quota, err := s.GetOrCreateUserQuota(ctx, userID)
	if err != nil {
		return nil, err
	}

	balance := quota.Quota - quota.UsedQuota
	affCode := ""
	if quota.AffCode != nil {
		affCode = *quota.AffCode
	}
	return &UserQuotaInfo{
		UserID:      quota.UserID,
		Quota:       quota.Quota,
		UsedQuota:   quota.UsedQuota,
		Group:       quota.Group,
		AffCode:     affCode,
		InviterID:   quota.InviterID,
		BalanceYuan: float64(balance) / 500000.0, // 转换为元
	}, nil
}

// AddQuota adds quota to user account (recharge)
func (s *UserQuotaService) AddQuota(ctx context.Context, userID int, amount int64, description, orderID string) error {
	_, err := authz.RunWithSystemBypass(ctx, "add-quota", func(bypassCtx context.Context) (struct{}, error) {
		client := s.entFromContext(bypassCtx)

		tx, err := client.Tx(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		// Get user quota
		quota, err := tx.UserQuota.Query().
			Where(userquota.UserIDEQ(userID)).
			Only(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to get user quota: %w", err)
		}

		// Update quota
		newQuota := quota.Quota + amount
		_, err = tx.UserQuota.UpdateOne(quota).
			SetQuota(newQuota).
			Save(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to update quota: %w", err)
		}

		// Create transaction record
		_, err = tx.QuotaTransaction.Create().
			SetUserID(userID).
			SetType(quotatransaction.TypeRecharge).
			SetAmount(amount).
			SetBalanceAfter(newQuota).
			SetNillableDescription(&description).
			SetNillableOrderID(&orderID).
			SetStatus(quotatransaction.StatusCompleted).
			Save(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to create transaction: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return struct{}{}, fmt.Errorf("failed to commit: %w", err)
		}

		log.Info(ctx, "quota added", 
			log.Int("user_id", userID), 
			log.Int64("amount", amount),
			log.String("order_id", orderID))

		return struct{}{}, nil
	})
	return err
}

// ConsumeQuota consumes quota from user account
func (s *UserQuotaService) ConsumeQuota(ctx context.Context, userID int, amount int64, description string) error {
	_, err := authz.RunWithSystemBypass(ctx, "consume-quota", func(bypassCtx context.Context) (struct{}, error) {
		client := s.entFromContext(bypassCtx)

		tx, err := client.Tx(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		quota, err := tx.UserQuota.Query().
			Where(userquota.UserIDEQ(userID)).
			Only(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to get user quota: %w", err)
		}

		// Check balance
		balance := quota.Quota - quota.UsedQuota
		if balance < amount {
			return struct{}{}, fmt.Errorf("insufficient balance: have %d, need %d", balance, amount)
		}

		// Update used quota
		newUsedQuota := quota.UsedQuota + amount
		_, err = tx.UserQuota.UpdateOne(quota).
			SetUsedQuota(newUsedQuota).
			Save(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to update quota: %w", err)
		}

		// Create transaction record
		_, err = tx.QuotaTransaction.Create().
			SetUserID(userID).
			SetType(quotatransaction.TypeConsume).
			SetAmount(-amount). // Negative for consumption
			SetBalanceAfter(quota.Quota - newUsedQuota).
			SetNillableDescription(&description).
			SetStatus(quotatransaction.StatusCompleted).
			Save(bypassCtx)
		if err != nil {
			return struct{}{}, fmt.Errorf("failed to create transaction: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return struct{}{}, fmt.Errorf("failed to commit: %w", err)
		}

		return struct{}{}, nil
	})
	return err
}

// GetTransactionHistory gets user transaction history
func (s *UserQuotaService) GetTransactionHistory(ctx context.Context, userID int, limit int) ([]*ent.QuotaTransaction, error) {
	return authz.RunWithSystemBypass(ctx, "get-transactions", func(bypassCtx context.Context) ([]*ent.QuotaTransaction, error) {
		client := s.entFromContext(bypassCtx)
		return client.QuotaTransaction.Query().
			Where(quotatransaction.UserIDEQ(userID)).
			Order(ent.Desc(quotatransaction.FieldCreatedAt)).
			Limit(limit).
			All(bypassCtx)
	})
}

// generateAffCode generates a random affiliate code
func generateAffCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ConvertQuotaToYuan converts quota (in fen) to yuan
func ConvertQuotaToYuan(quota int64) float64 {
	return float64(quota) / 500000.0
}

// ConvertYuanToQuota converts yuan to quota (in fen)
func ConvertYuanToQuota(yuan float64) int64 {
	return int64(yuan * 500000)
}
