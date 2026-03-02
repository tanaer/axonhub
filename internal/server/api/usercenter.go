package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/apikey"
	"github.com/looplj/axonhub/internal/ent/rechargeorder"
	"github.com/looplj/axonhub/internal/ent/userquota"
	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/server/biz"
)

type UserCenterHandlersParams struct {
	fx.In

	Client           *ent.Client
	AuthService      *biz.AuthService      `optional:"true"`
	UserQuotaService *biz.UserQuotaService `optional:"true"`
}

func NewUserCenterHandlers(params UserCenterHandlersParams) *UserCenterHandlers {
	return &UserCenterHandlers{
		Client:           params.Client,
		AuthService:      params.AuthService,
		UserQuotaService: params.UserQuotaService,
	}
}

type UserCenterHandlers struct {
	Client           *ent.Client
	AuthService      *biz.AuthService
	UserQuotaService *biz.UserQuotaService
}

// APIKeyResponse represents an API key in responses
type APIKeyResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"`
	Prefix    string `json:"prefix"`
	CreatedAt string `json:"created_at"`
	IsActive  bool   `json:"is_active"`
}

// CreateAPIKeyRequest represents the request to create an API key
type CreateAPIKeyRequest struct {
	Name string `json:"name" binding:"required"`
}

// getUserIDFromContext extracts user ID from the context
func getUserIDFromContext(c *gin.Context) (int, error) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		return 0, errors.New("unauthorized")
	}
	return user.ID, nil
}

// GetAPIKeys returns all API keys for the current user
func (h *UserCenterHandlers) GetAPIKeys(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// Use system bypass to query user's API keys
	keys, err := authz.RunWithSystemBypass(ctx, "get-apikeys", func(bypassCtx context.Context) ([]*ent.APIKey, error) {
		return h.Client.APIKey.Query().
			Where(apikey.UserIDEQ(userID)).
			Order(ent.Desc(apikey.FieldCreatedAt)).
			All(bypassCtx)
	})
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to fetch API keys"))
		return
	}

	response := make([]APIKeyResponse, len(keys))
	for i, k := range keys {
		prefix := ""
		if len(k.Key) >= 8 {
			prefix = k.Key[:8]
		}
		response[i] = APIKeyResponse{
			ID:        k.ID,
			Name:      k.Name,
			Prefix:    prefix,
			CreatedAt: k.CreatedAt.Format(time.RFC3339),
			IsActive:  k.Status == apikey.StatusEnabled,
		}
	}

	c.JSON(http.StatusOK, response)
}

// CreateAPIKey creates a new API key for the current user
func (h *UserCenterHandlers) CreateAPIKey(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	// Generate API key
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to generate key"))
		return
	}
	apiKeyStr := "sk-" + hex.EncodeToString(keyBytes)

	// Create API key with system bypass
	newKey, err := authz.RunWithSystemBypass(ctx, "create-apikey", func(bypassCtx context.Context) (*ent.APIKey, error) {
		return h.Client.APIKey.Create().
			SetName(req.Name).
			SetKey(apiKeyStr).
			SetUserID(userID).
			SetStatus(apikey.StatusEnabled).
			SetType(apikey.TypeUser).
			Save(bypassCtx)
	})
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to create API key: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, APIKeyResponse{
		ID:        newKey.ID,
		Name:      newKey.Name,
		Key:       apiKeyStr,
		Prefix:    apiKeyStr[:8],
		CreatedAt: newKey.CreatedAt.Format(time.RFC3339),
		IsActive:  true,
	})
}

// DeleteAPIKey deletes an API key
func (h *UserCenterHandlers) DeleteAPIKey(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid key ID"))
		return
	}

	// Delete with system bypass, but only if it belongs to the user
	err = authz.RunWithSystemBypassVoid(ctx, "delete-apikey", func(bypassCtx context.Context) error {
		deleted, err := h.Client.APIKey.Delete().
			Where(
				apikey.IDEQ(keyID),
				apikey.UserIDEQ(userID),
			).
			Exec(bypassCtx)
		if err != nil {
			return err
		}
		if deleted == 0 {
			return errors.New("API key not found")
		}
		return nil
	})
	if err != nil {
		JSONError(c, http.StatusNotFound, errors.New("API key not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ========== 订单管理 ==========

// OrderResponse represents an order in responses
type OrderResponse struct {
	ID           int     `json:"id"`
	OrderID      string  `json:"order_no"`
	Type         string  `json:"type"`
	Amount       int64   `json:"amount"`
	AmountYuan   float64 `json:"amount_yuan"`
	Status       string  `json:"status"`
	PaymentMethod string `json:"payment_method,omitempty"`
	PaidAt       string  `json:"paid_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
	Description  string  `json:"description,omitempty"`
}

// GetUserOrders returns the user's order history
func (h *UserCenterHandlers) GetUserOrders(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	orders, err := authz.RunWithSystemBypass(ctx, "get-orders", func(bypassCtx context.Context) ([]*ent.RechargeOrder, error) {
		return h.Client.RechargeOrder.Query().
			Where(rechargeorder.UserIDEQ(userID)).
			Order(ent.Desc(rechargeorder.FieldCreatedAt)).
			Limit(limit).
			All(bypassCtx)
	})
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to fetch orders"))
		return
	}

	response := make([]OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = OrderResponse{
			ID:            o.ID,
			OrderID:       o.OrderID,
			Type:          string(o.Type),
			Amount:        o.Amount,
			AmountYuan:    float64(o.Amount) / 100.0,
			Status:        string(o.Status),
			PaymentMethod: o.PaymentMethod,
			CreatedAt:     o.CreatedAt.Format(time.RFC3339),
			Description:   o.Description,
		}
		if o.PaidAt != nil {
			response[i].PaidAt = o.PaidAt.Format(time.RFC3339)
		}
	}

	c.JSON(http.StatusOK, gin.H{"orders": response})
}

// ========== 推荐系统 ==========

// ReferralInfoResponse represents referral info
type ReferralInfoResponse struct {
	Code           string  `json:"code"`
	Link           string  `json:"link"`
	TotalReferrals int     `json:"total_referrals"`
	TotalRewards   float64 `json:"total_rewards"`
}

// GetReferralInfo returns the user's referral info
func (h *UserCenterHandlers) GetReferralInfo(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// 获取用户推荐信息
	quota, err := authz.RunWithSystemBypass(ctx, "get-referral-info", func(bypassCtx context.Context) (*ent.UserQuota, error) {
		return h.Client.UserQuota.Query().
			Where(userquota.UserIDEQ(userID)).
			Only(bypassCtx)
	})
	if err != nil && !ent.IsNotFound(err) {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get referral info"))
		return
	}

	// 如果没有推荐码，创建一个
	var affCode string
	if quota != nil && quota.AffCode != nil {
		affCode = *quota.AffCode
	} else {
		// 生成推荐码
		affCode = generateReferralCode(userID)
		// 更新用户推荐码
		_, err = authz.RunWithSystemBypass(ctx, "update-aff-code", func(bypassCtx context.Context) (*ent.UserQuota, error) {
			return h.Client.UserQuota.Create().
				SetUserID(userID).
				SetAffCode(affCode).
				SetQuota(0).
				SetUsedQuota(0).
				Save(bypassCtx)
		})
		if err != nil {
			// 可能已存在，尝试更新
			authz.RunWithSystemBypassVoid(ctx, "update-aff-code-2", func(bypassCtx context.Context) error {
				_, err := h.Client.UserQuota.Update().
					Where(userquota.UserIDEQ(userID)).
					SetAffCode(affCode).
					Save(bypassCtx)
				return err
			})
		}
	}

	// 统计推荐人数
	totalReferrals := 0
	if h.UserQuotaService != nil {
		referrals, _ := authz.RunWithSystemBypass(ctx, "count-referrals", func(bypassCtx context.Context) (int, error) {
			return h.Client.UserQuota.Query().
				Where(userquota.InviterIDEQ(userID)).
				Count(bypassCtx)
		})
		totalReferrals = referrals
	}

	// 计算推荐奖励（从交易记录中统计）
	var totalRewards int64 = 0
	authz.RunWithSystemBypassVoid(ctx, "get-rewards", func(bypassCtx context.Context) error {
		// TODO: 从 QuotaTransaction 表中统计推荐奖励
		return nil
	})

	// 推荐链接
	link := fmt.Sprintf("https://dev.claudeai.best/register?ref=%s", affCode)

	c.JSON(http.StatusOK, ReferralInfoResponse{
		Code:           affCode,
		Link:           link,
		TotalReferrals: totalReferrals,
		TotalRewards:   float64(totalRewards) / 100.0,
	})
}

// ReferralRecord represents a referral record
type ReferralRecord struct {
	ID           int     `json:"id"`
	ReferredUser string  `json:"referred_email"`
	RewardAmount float64 `json:"reward_amount"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// GetReferralRecords returns the user's referral records
func (h *UserCenterHandlers) GetReferralRecords(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// 获取被推荐的用户列表
	quotas, err := authz.RunWithSystemBypass(ctx, "get-referrals", func(bypassCtx context.Context) ([]*ent.UserQuota, error) {
		return h.Client.UserQuota.Query().
			Where(userquota.InviterIDEQ(userID)).
			Order(ent.Desc(userquota.FieldCreatedAt)).
			Limit(50).
			All(bypassCtx)
	})
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get referral records"))
		return
	}

	records := make([]ReferralRecord, len(quotas))
	for i, q := range quotas {
		email := ""
		if false {
			email = ""
		}
		createdAt := ""
		if !q.CreatedAt.IsZero() {
			createdAt = q.CreatedAt.Format(time.RFC3339)
		}
		records[i] = ReferralRecord{
			ID:           q.ID,
			ReferredUser: email,
			RewardAmount: 0, // TODO: 从交易记录中获取奖励金额
			Status:       "completed",
			CreatedAt:    createdAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

// GetUserReferral returns the user's referral information (legacy endpoint)
func (h *UserCenterHandlers) GetUserReferral(c *gin.Context) {
	h.GetReferralInfo(c)
}

// generateReferralCode generates a unique referral code
func generateReferralCode(userID int) string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("REF%s%s", strconv.Itoa(userID), hex.EncodeToString(b))
}

// ========== 账户管理 ==========

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=7"`
}

// ChangePassword handles password change requests
func (h *UserCenterHandlers) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request: "+err.Error()))
		return
	}

	// Get the user
	user, err := h.Client.User.Get(ctx, userID)
	if err != nil {
		JSONError(c, http.StatusNotFound, errors.New("User not found"))
		return
	}

	// Verify current password using auth service
	if h.AuthService != nil {
		_, err := h.AuthService.AuthenticateUser(ctx, user.Email, req.CurrentPassword)
		if err != nil {
			JSONError(c, http.StatusBadRequest, errors.New("Current password is incorrect"))
			return
		}
	}

	// Update password
	_, err = h.Client.User.UpdateOneID(userID).
		SetPassword(req.NewPassword).
		Save(ctx)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to update password: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// UpdateProfileRequest represents the request to update profile
type UpdateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UpdateProfile handles profile update requests
func (h *UserCenterHandlers) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	update := h.Client.User.UpdateOneID(userID)
	if req.FirstName != "" {
		update.SetFirstName(req.FirstName)
	}
	if req.LastName != "" {
		update.SetLastName(req.LastName)
	}

	_, err = update.Save(ctx)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to update profile"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile updated successfully",
	})
}

// ========== 余额扣除和检查 ==========

// CheckBalanceRequest represents a balance check request
type CheckBalanceRequest struct {
	Amount int64 `json:"amount" binding:"required"`
}

// CheckBalance checks if user has enough balance
func (h *UserCenterHandlers) CheckBalance(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	var req CheckBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	// 获取用户余额
	quota, err := authz.RunWithSystemBypass(ctx, "check-balance", func(bypassCtx context.Context) (*ent.UserQuota, error) {
		return h.Client.UserQuota.Query().
			Where(userquota.UserIDEQ(userID)).
			Only(bypassCtx)
	})
	if err != nil {
		if ent.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{
				"sufficient": false,
				"balance":    0,
				"required":   req.Amount,
			})
			return
		}
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to check balance"))
		return
	}

	balance := quota.Quota - quota.UsedQuota
	sufficient := balance >= req.Amount

	c.JSON(http.StatusOK, gin.H{
		"sufficient": sufficient,
		"balance":    balance,
		"balance_yuan": float64(balance) / 100.0,
		"required":   req.Amount,
		"required_yuan": float64(req.Amount) / 100.0,
	})
}

// ConsumeBalance consumes balance from user account
func (h *UserCenterHandlers) ConsumeBalance(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	var req struct {
		Amount      int64  `json:"amount" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	if h.UserQuotaService == nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Quota service not available"))
		return
	}

	// 扣除余额
	err = h.UserQuotaService.ConsumeQuota(ctx, userID, req.Amount, req.Description)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	log.Info(ctx, "balance consumed",
		log.Int("user_id", userID),
		log.Int64("amount", req.Amount),
		log.String("description", req.Description))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Balance consumed successfully",
		"amount":  req.Amount,
	})
}

// ========== API Key 余额检查中间件 ==========

// CheckAPIKeyBalance checks if an API key has sufficient balance
func (h *UserCenterHandlers) CheckAPIKeyBalance(ctx context.Context, apiKey string, requiredAmount int64) (bool, int64, error) {
	// 获取 API Key 关联的用户
	key, err := authz.RunWithSystemBypass(ctx, "get-apikey-user", func(bypassCtx context.Context) (*ent.APIKey, error) {
		return h.Client.APIKey.Query().
			Where(apikey.KeyEQ(apiKey)).
			Only(bypassCtx)
	})
	if err != nil {
		return false, 0, err
	}

	// 获取用户余额
	quota, err := authz.RunWithSystemBypass(ctx, "get-user-quota", func(bypassCtx context.Context) (*ent.UserQuota, error) {
		return h.Client.UserQuota.Query().
			Where(userquota.UserIDEQ(key.UserID)).
			Only(bypassCtx)
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return false, 0, nil
		}
		return false, 0, err
	}

	balance := quota.Quota - quota.UsedQuota
	return balance >= requiredAmount, balance, nil
}

// ConsumeAPIKeyBalance consumes balance for an API key call
func (h *UserCenterHandlers) ConsumeAPIKeyBalance(ctx context.Context, apiKey string, amount int64, description string) error {
	// 获取 API Key 关联的用户
	key, err := authz.RunWithSystemBypass(ctx, "get-apikey-for-consume", func(bypassCtx context.Context) (*ent.APIKey, error) {
		return h.Client.APIKey.Query().
			Where(apikey.KeyEQ(apiKey)).
			Only(bypassCtx)
	})
	if err != nil {
		return err
	}

	// 使用 UserQuotaService 扣除余额
	if h.UserQuotaService == nil {
		return errors.New("quota service not available")
	}

	return h.UserQuotaService.ConsumeQuota(ctx, key.UserID, amount, description)
}
