package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/server/biz"
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

// EPUSDT API 配置
const (
	EPUSDTBaseURL = "http://localhost:8080"
	EPUSDTToken   = "649686E56BB62262B4E327C229D836D3"
)

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
	Amount        int64  `json:"amount" binding:"required,min=100"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// Recharge creates a recharge order
func (h *UserQuotaHandler) Recharge(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	var req RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request: "+err.Error()))
		return
	}

	orderID := "RCH" + time.Now().Format("20060102150405") + strconv.Itoa(user.ID)

	if req.PaymentMethod == "epusdt" {
		// 创建 EPUSDT 支付订单
		paymentURL, err := h.createEPUSDTOrder(orderID, req.Amount)
		if err != nil {
			JSONError(c, http.StatusInternalServerError, errors.New("Failed to create payment order: "+err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"order_id":       orderID,
			"amount":         req.Amount,
			"amount_yuan":    float64(req.Amount) / 100.0,
			"payment_method": req.PaymentMethod,
			"payment_url":    paymentURL,
			"status":         "pending",
		})
		return
	}

	// Stripe 或其他支付方式
	c.JSON(http.StatusOK, gin.H{
		"order_id":       orderID,
		"amount":         req.Amount,
		"amount_yuan":    float64(req.Amount) / 100.0,
		"payment_method": req.PaymentMethod,
		"status":         "pending",
		"message":        "Payment integration pending",
	})
}

// EPUSDT 创建订单请求
type EPUSDTOrderRequest struct {
	OrderID   string `json:"order_id"`
	Amount    string `json:"amount"`
	NotifyURL string `json:"notify_url"`
	ReturnURL string `json:"return_url"`
}

// EPUSDT 创建订单响应
type EPUSDTOrderResponse struct {
	TradeID   string `json:"trade_id"`
	OrderID   string `json:"order_id"`
	Amount    string `json:"amount"`
	PaymentURL string `json:"payment_url"`
	ExpiresAt string `json:"expires_at"`
}

func (h *UserQuotaHandler) createEPUSDTOrder(orderID string, amount int64) (string, error) {
	// TODO: EPUSDT API 需要进一步配置
	// 暂时返回模拟支付页面 URL
	amountUSDT := float64(amount) / 100.0 / 7.0
	paymentURL := fmt.Sprintf("%s/payment?order_id=%s&amount=%.2f", EPUSDTBaseURL, orderID, amountUSDT)
	return paymentURL, nil
}

// EPUSDTCallback 处理 EPUSDT 支付回调
func (h *UserQuotaHandler) EPUSDTCallback(c *gin.Context) {
	var callback struct {
		TradeID  string `json:"trade_id"`
		OrderID  string `json:"order_id"`
		Amount   string `json:"amount"`
		Status   int    `json:"status"`
		Sign     string `json:"sign"`
	}
	
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request"})
		return
	}
	
	// 验证签名 (TODO: 实现签名验证)
	
	// 如果支付成功，更新用户余额
	if callback.Status == 2 { // 2 表示支付成功
		amount, _ := strconv.ParseFloat(callback.Amount, 64)
		amountFen := int64(amount * 7 * 100) // USDT 转 CNY 再转分
		
		// 从订单号解析用户ID
		// 订单格式: RCH20060102150405123
		if len(callback.OrderID) > 17 {
			userIDStr := callback.OrderID[17:]
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
				return
			}
			
			err = h.UserQuotaService.AddQuota(c.Request.Context(), userID, amountFen, "EPUSDT 充值", callback.OrderID)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
				return
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

func generateOrderID() string {
	return "ORD-" + time.Now().Format("20060102150405")
}
