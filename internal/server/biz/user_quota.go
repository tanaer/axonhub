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

// ========== Admin Statistics ==========

// AdminOverview represents platform-wide statistics
type AdminOverview struct {
	TotalUsers       int64   `json:"total_users"`
	TodayUsers       int64   `json:"today_users"`
	TotalRevenue     float64 `json:"total_revenue"`     // Total revenue in yuan
	TodayRevenue     float64 `json:"today_revenue"`     // Today's revenue in yuan
	ActiveSubs       int64   `json:"active_subscriptions"`
	TotalTransactions int64  `json:"total_transactions"`
}

// DailyCount represents daily statistics
type DailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DailyRevenue represents daily revenue
type DailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

// GetAdminOverview returns platform-wide statistics
func (s *UserQuotaService) GetAdminOverview(ctx context.Context) (*AdminOverview, error) {
	return authz.RunWithSystemBypass(ctx, "admin-overview", func(bypassCtx context.Context) (*AdminOverview, error) {
		client := s.entFromContext(bypassCtx)
		overview := &AdminOverview{}

		// Total users
		totalUsers, err := client.User.Query().Count(bypassCtx)
		if err != nil {
			log.Warn(bypassCtx, "Failed to count users", log.Cause(err))
		}
		overview.TotalUsers = int64(totalUsers)

		// Total transactions
		totalTx, err := client.QuotaTransaction.Query().Count(bypassCtx)
		if err != nil {
			log.Warn(bypassCtx, "Failed to count transactions", log.Cause(err))
		}
		overview.TotalTransactions = int64(totalTx)

		// Total revenue (sum of recharge transactions)
		transactions, err := client.QuotaTransaction.Query().
			Where(quotatransaction.TypeEQ("recharge")).
			All(bypassCtx)
		if err == nil {
			var total int64
			for _, tx := range transactions {
				total += tx.Amount
			}
			overview.TotalRevenue = float64(total) / 100.0 // Convert from cents
		}

		return overview, nil
	})
}

// GetUserRegistrationTrend returns user registration trend for last N days
func (s *UserQuotaService) GetUserRegistrationTrend(ctx context.Context, days int) ([]DailyCount, error) {
	return authz.RunWithSystemBypass(ctx, "user-trend", func(bypassCtx context.Context) ([]DailyCount, error) {
		client := s.entFromContext(bypassCtx)
		
		// Get all users created in last N days
		users, err := client.User.Query().All(bypassCtx)
		if err != nil {
			return nil, err
		}

		// Group by date
		counts := make(map[string]int64)
		for _, u := range users {
			date := u.CreatedAt.Format("2006-01-02")
			counts[date]++
		}

		// Convert to slice
		var result []DailyCount
		for date, count := range counts {
			result = append(result, DailyCount{Date: date, Count: count})
		}

		return result, nil
	})
}

// GetRevenueTrend returns revenue trend for last N days
func (s *UserQuotaService) GetRevenueTrend(ctx context.Context, days int) ([]DailyRevenue, error) {
	return authz.RunWithSystemBypass(ctx, "revenue-trend", func(bypassCtx context.Context) ([]DailyRevenue, error) {
		client := s.entFromContext(bypassCtx)
		
		// Get all recharge transactions
		transactions, err := client.QuotaTransaction.Query().
			Where(quotatransaction.TypeEQ("recharge")).
			All(bypassCtx)
		if err != nil {
			return nil, err
		}

		// Group by date
		revenues := make(map[string]float64)
		for _, tx := range transactions {
			date := tx.CreatedAt.Format("2006-01-02")
			revenues[date] += float64(tx.Amount) / 100.0
		}

		// Convert to slice
		var result []DailyRevenue
		for date, revenue := range revenues {
			result = append(result, DailyRevenue{Date: date, Revenue: revenue})
		}

		return result, nil
	})
}
