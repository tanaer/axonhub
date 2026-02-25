package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/server/biz"
	"go.uber.org/fx"
)

type UserQuotaHandlerParams struct {
	fx.In

	UserQuotaService *biz.UserQuotaService
}

func NewUserQuotaHandler(params UserQuotaHandlerParams) *UserQuotaHandler {
	return &UserQuotaHandler{
		UserQuotaService: params.UserQuotaService,
	}
}

type UserQuotaHandler struct {
	UserQuotaService *biz.UserQuotaService
}

// GetMyQuota returns current user's quota info
func (h *UserQuotaHandler) GetMyQuota(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	info, err := h.UserQuotaService.GetUserQuotaInfo(c.Request.Context(), user.ID)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get quota: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetTransactionHistory returns user's transaction history
func (h *UserQuotaHandler) GetTransactionHistory(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	transactions, err := h.UserQuotaService.GetTransactionHistory(c.Request.Context(), user.ID, 50)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get transactions: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}

// RechargeRequest represents recharge request
type RechargeRequest struct {
	Amount        int64  `json:"amount" binding:"required,min=1"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// Recharge creates a recharge order
func (h *UserQuotaHandler) Recharge(c *gin.Context) {
	_, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request: "+err.Error()))
		return
	}

	// TODO: Integrate with Stripe/EPUSDT payment gateway
	// For now, just create a pending order
	orderID := generateOrderID()

	c.JSON(http.StatusOK, gin.H{
		"order_id":       orderID,
		"amount":         req.Amount,
		"payment_method": req.PaymentMethod,
		"status":         "pending",
		"message":        "Payment integration pending",
	})
}

func generateOrderID() string {
	return "ORD-" + time.Now().Format("20060102150405")
}
