package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/apikey"
	"github.com/looplj/axonhub/internal/server/biz"
)

type UserCenterHandlersParams struct {
	fx.In

	Client      *ent.Client
	AuthService *biz.AuthService `optional:"true"`
}

func NewUserCenterHandlers(params UserCenterHandlersParams) *UserCenterHandlers {
	return &UserCenterHandlers{
		Client:      params.Client,
		AuthService: params.AuthService,
	}
}

type UserCenterHandlers struct {
	Client      *ent.Client
	AuthService *biz.AuthService
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

// GetUserOrders returns the user's order history
func (h *UserCenterHandlers) GetUserOrders(c *gin.Context) {
	_, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// TODO: Implement actual order lookup
	c.JSON(http.StatusOK, []interface{}{})
}

// GetUserReferral returns the user's referral information
func (h *UserCenterHandlers) GetUserReferral(c *gin.Context) {
	_, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// TODO: Implement actual referral lookup
	c.JSON(http.StatusOK, gin.H{
		"referral_code": "",
		"earnings":      0,
		"referrals":     0,
	})
}

// GetReferralInfo returns the user's referral info
func (h *UserCenterHandlers) GetReferralInfo(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// TODO: Implement actual referral lookup from database
	// For now, return placeholder data
	c.JSON(http.StatusOK, gin.H{
		"referral_code":  "REF" + strconv.Itoa(userID),
		"total_earnings": 0,
		"total_referrals": 0,
		"pending_rewards": 0,
	})
}

// GetReferralRecords returns the user's referral records
func (h *UserCenterHandlers) GetReferralRecords(c *gin.Context) {
	_, err := getUserIDFromContext(c)
	if err != nil {
		JSONError(c, http.StatusUnauthorized, err)
		return
	}

	// TODO: Implement actual referral records lookup from database
	// For now, return empty array
	c.JSON(http.StatusOK, []interface{}{})
}

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
