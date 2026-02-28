package api

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/log"
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

// EPUSDT API 配置 - 从环境变量读取，开发环境使用模拟模式
var (
	EPUSDTBaseURL = getEnvOrDefault("EPUSDT_BASE_URL", "")
	EPUSDTToken   = getEnvOrDefault("EPUSDT_TOKEN", "649686E56BB62262B4E327C229D836D3")
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

// GetPackages returns available subscription packages
func (h *UserQuotaHandler) GetPackages(c *gin.Context) {
	// TODO: 从数据库获取套餐列表
	// 暂时返回预设套餐
	packages := []gin.H{
		{
			"id":            1,
			"name":          "starter",
			"display_name":  "体验版",
			"price":         0,
			"price_yuan":    0,
			"currency":      "CNY",
			"quota":         100000,  // 10元
			"quota_yuan":    10,
			"duration_days": 30,
			"features":      []string{"免费试用", "基础模型", "社区支持"},
			"is_popular":    false,
		},
		{
			"id":            2,
			"name":          "basic",
			"display_name":  "基础版",
			"price":         9900,    // 99元
			"price_yuan":    99,
			"currency":      "CNY",
			"quota":         1000000, // 100元
			"quota_yuan":    100,
			"duration_days": 30,
			"features":      []string{"5 个 API Key", "全模型接入", "邮件支持"},
			"is_popular":    false,
		},
		{
			"id":            3,
			"name":          "pro",
			"display_name":  "专业版",
			"price":         29900,   // 299元
			"price_yuan":    299,
			"currency":      "CNY",
			"quota":         5000000, // 500元
			"quota_yuan":    500,
			"duration_days": 30,
			"features":      []string{"无限 API Key", "优先支持", "用量分析"},
			"is_popular":    true,
		},
		{
			"id":            4,
			"name":          "enterprise",
			"display_name":  "企业版",
			"price":         99900,   // 999元
			"price_yuan":    999,
			"currency":      "CNY",
			"quota":         -1,      // 无限制
			"quota_yuan":    -1,
			"duration_days": 30,
			"features":      []string{"专属部署", "SLA 保障", "7×24 支持"},
			"is_popular":    false,
		},
	}

	c.JSON(http.StatusOK, gin.H{"packages": packages})
}

// PurchasePackageRequest represents package purchase request
type PurchasePackageRequest struct {
	PackageID     int    `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// PackageInfo represents a package
type PackageInfo struct {
	ID          int
	Price       int
	Quota       int
	DisplayName string
}

// PurchasePackage creates a package purchase order
func (h *UserQuotaHandler) PurchasePackage(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	var req PurchasePackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request: "+err.Error()))
		return
	}

	// 获取套餐信息
	packages := []PackageInfo{
		{ID: 1, Price: 0, Quota: 100000, DisplayName: "体验版"},
		{ID: 2, Price: 9900, Quota: 1000000, DisplayName: "基础版"},
		{ID: 3, Price: 29900, Quota: 5000000, DisplayName: "专业版"},
		{ID: 4, Price: 99900, Quota: -1, DisplayName: "企业版"},
	}
	
	var pkg *PackageInfo
	for i := range packages {
		if packages[i].ID == req.PackageID {
			pkg = &packages[i]
			break
		}
	}
	if pkg == nil {
		JSONError(c, http.StatusBadRequest, errors.New("Package not found"))
		return
	}

	// 如果是免费套餐，直接发放
	if pkg.Price == 0 {
		err := h.UserQuotaService.AddQuota(c.Request.Context(), user.ID, int64(pkg.Quota), "领取免费套餐: "+pkg.DisplayName, "FREE-"+time.Now().Format("20060102150405"))
		if err != nil {
			JSONError(c, http.StatusInternalServerError, errors.New("Failed to add quota: "+err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"message":     "套餐已激活",
			"quota":       pkg.Quota,
			"quota_yuan":  float64(pkg.Quota) / 100.0,
		})
		return
	}

	// 付费套餐，创建支付订单
	orderID := "PKG" + time.Now().Format("20060102150405") + strconv.Itoa(user.ID)

	if req.PaymentMethod == "epusdt" {
		paymentURL, err := h.createEPUSDTOrder(orderID, int64(pkg.Price))
		if err != nil {
			JSONError(c, http.StatusInternalServerError, errors.New("Failed to create payment order: "+err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"order_id":     orderID,
			"package_id":   req.PackageID,
			"package_name": pkg.DisplayName,
			"amount":       pkg.Price,
			"amount_yuan":  float64(pkg.Price) / 100.0,
			"quota":        pkg.Quota,
			"payment_url":  paymentURL,
			"status":       "pending",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"order_id":     orderID,
		"package_id":   req.PackageID,
		"package_name": pkg.DisplayName,
		"amount":       pkg.Price,
		"amount_yuan":  float64(pkg.Price) / 100.0,
		"status":       "pending",
	})
}
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
	// 如果没有配置 EPUSDT URL，使用模拟支付页面
	if EPUSDTBaseURL == "" {
		// 返回前端模拟支付页面 URL
		amountUSDT := float64(amount) / 100.0 / 7.0
		return fmt.Sprintf("/user-center/payment?order_id=%s&amount=%.2f&method=epusdt", orderID, amountUSDT), nil
	}
	
	amountUSDT := float64(amount) / 100.0 / 7.0
	paymentURL := fmt.Sprintf("%s/payment?order_id=%s&amount=%.2f", EPUSDTBaseURL, orderID, amountUSDT)
	return paymentURL, nil
}

// MockPaymentConfirm 模拟支付确认（开发/测试环境使用）
func (h *UserQuotaHandler) MockPaymentConfirm(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
		Amount  int64  `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	// 从订单号解析用户ID
	orderID := req.OrderID
	if len(orderID) < 17 {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid order ID"))
		return
	}

	userIDStr := orderID[17:]
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid order ID format"))
		return
	}

	// 添加用户余额
	description := "模拟充值（测试环境）"
	err = h.UserQuotaService.AddQuota(c.Request.Context(), userID, req.Amount, description, orderID)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to add quota: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Payment confirmed",
		"order_id":     orderID,
		"amount":       req.Amount,
		"amount_yuan":  float64(req.Amount) / 100.0,
	})
}

// EPUSDTCallback 处理 EPUSDT 支付回调
func (h *UserQuotaHandler) EPUSDTCallback(c *gin.Context) {
	var callback struct {
		TradeID    string `json:"trade_id"`
		OrderID    string `json:"order_id"`
		Amount     string `json:"amount"`
		ActualAmount string `json:"actual_amount"`
		Token      string `json:"token"`
		Status     int    `json:"status"`
		Sign       string `json:"sign"`
		Timestamp  int64  `json:"timestamp"`
	}
	
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request"})
		return
	}
	
	// 验证签名
	if !verifyEPUSDTSign(callback.TradeID, callback.OrderID, callback.Amount, callback.Status, callback.Sign) {
		c.JSON(http.StatusOK, gin.H{"code": 403, "message": "Invalid signature"})
		return
	}
	
	// 如果支付成功，更新用户余额
	if callback.Status == 2 { // 2 表示支付成功
		amount, _ := strconv.ParseFloat(callback.Amount, 64)
		amountFen := int64(amount * 7 * 100) // USDT 转 CNY 再转分
		
		// 从订单号解析用户ID
		// 订单格式: RCH20060102150405123 或 PKG20060102150405123
		orderID := callback.OrderID
		var userID int
		var err error
		
		if len(orderID) > 17 {
			userIDStr := orderID[17:]
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
				return
			}
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
			return
		}
		
		// 添加用户余额
		description := "EPUSDT 充值"
		if len(orderID) > 3 && orderID[:3] == "PKG" {
			description = "套餐购买"
		}
		
		err = h.UserQuotaService.AddQuota(c.Request.Context(), userID, amountFen, description, orderID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// verifyEPUSDTSign 验证 EPUSDT 回调签名
func verifyEPUSDTSign(tradeID, orderID, amount string, status int, sign string) bool {
	// EPUSDT 签名格式: MD5(trade_id + order_id + amount + status + secret)
	// 或者 HMAC-SHA256
	
	// 如果签名为空，暂时跳过验证（开发环境）
	if sign == "" {
		return true
	}
	
	// 方式1: MD5 签名
	expectedSign := md5.Sum([]byte(tradeID + orderID + amount + strconv.Itoa(status) + EPUSDTToken))
	expectedSignStr := hex.EncodeToString(expectedSign[:])
	
	if hmac.Equal([]byte(sign), []byte(expectedSignStr)) {
		return true
	}
	
	// 方式2: HMAC-SHA256 签名
	mac := hmac.New(sha256.New, []byte(EPUSDTToken))
	mac.Write([]byte(tradeID + orderID + amount + strconv.Itoa(status)))
	expectedHMAC := hex.EncodeToString(mac.Sum(nil))
	
	return hmac.Equal([]byte(sign), []byte(expectedHMAC))
}

func generateOrderID() string {
	return "ORD-" + time.Now().Format("20060102150405")
}

// ========== Admin Statistics API ==========

// GetAdminOverview returns overall platform statistics (admin only)
func (h *UserQuotaHandler) GetAdminOverview(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	// Log access for audit
	log.Info(c.Request.Context(), "Stats overview accessed", 
		log.Int("user_id", user.ID),
		log.String("email", user.Email))

	stats, err := h.UserQuotaService.GetAdminOverview(c.Request.Context())
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get stats: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetUserTrend returns user registration trend (admin only)
func (h *UserQuotaHandler) GetUserTrend(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	log.Info(c.Request.Context(), "User trend accessed", log.Int("user_id", user.ID))

	days := 7
	trend, err := h.UserQuotaService.GetUserRegistrationTrend(c.Request.Context(), days)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get trend: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"trend": trend})
}

// GetRevenueTrend returns revenue trend (admin only)
func (h *UserQuotaHandler) GetRevenueTrend(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	log.Info(c.Request.Context(), "Revenue trend accessed", log.Int("user_id", user.ID))

	days := 7
	trend, err := h.UserQuotaService.GetRevenueTrend(c.Request.Context(), days)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to get trend: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"trend": trend})
}
