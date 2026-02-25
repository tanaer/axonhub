package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"

	"github.com/looplj/axonhub/internal/contexts"
)

// StripeConfig holds Stripe configuration
type StripeConfig struct {
	SecretKey      string
	WebhookSecret  string
	SuccessURL     string
	CancelURL      string
	FrontendURL    string
}

// getStripeConfig returns Stripe configuration from environment
func getStripeConfig() *StripeConfig {
	return &StripeConfig{
		SecretKey:     os.Getenv("STRIPE_SECRET_KEY"),
		WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		SuccessURL:    os.Getenv("STRIPE_SUCCESS_URL"),
		CancelURL:     os.Getenv("STRIPE_CANCEL_URL"),
		FrontendURL:   getEnvOrDefault("FRONTEND_URL", "https://dev.claudeai.best"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// StripeCreateCheckoutSession creates a Stripe checkout session
func (h *UserQuotaHandler) StripeCreateCheckoutSession(c *gin.Context) {
	user, ok := contexts.GetUser(c.Request.Context())
	if !ok {
		JSONError(c, http.StatusUnauthorized, errors.New("Not authenticated"))
		return
	}

	var req struct {
		Amount        int64  `json:"amount" binding:"required,min=100"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request"))
		return
	}

	config := getStripeConfig()
	if config.SecretKey == "" {
		JSONError(c, http.StatusInternalServerError, errors.New("Stripe not configured"))
		return
	}

	stripe.Key = config.SecretKey

	// 金额转换为美元分 (假设汇率 1 USD = 7 CNY)
	amountUSD := req.Amount * 100 / 700 // 分 -> 美元分
	if amountUSD < 50 {
		amountUSD = 50 // Stripe 最低 $0.50
	}

	orderID := fmt.Sprintf("STRIPE%d%d", user.ID, time.Now().Unix())

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("usd"),
					UnitAmount: stripe.Int64(amountUSD),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String("MuskAPI Balance Recharge"),
						Description: stripe.String(fmt.Sprintf("Recharge ¥%.2f to your account", float64(req.Amount)/100)),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(config.FrontendURL + "/user-center/billing?payment=success"),
		CancelURL:  stripe.String(config.FrontendURL + "/user-center/billing?payment=cancelled"),
		Metadata: map[string]string{
			"user_id":    strconv.Itoa(user.ID),
			"order_id":   orderID,
			"amount_cny": strconv.FormatInt(req.Amount, 10),
		},
	}

	s, err := session.New(params)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Failed to create checkout session: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"session_id":   s.ID,
		"checkout_url": s.URL,
		"order_id":     orderID,
		"amount":       req.Amount,
		"amount_yuan":  float64(req.Amount) / 100.0,
	})
}

// StripeWebhook handles Stripe webhook events
func (h *UserQuotaHandler) StripeWebhook(c *gin.Context) {
	config := getStripeConfig()
	if config.WebhookSecret == "" {
		c.JSON(http.StatusOK, gin.H{"received": true})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	event, err := webhook.ConstructEvent(body, c.GetHeader("Stripe-Signature"), config.WebhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}

		// 获取元数据
		userIDStr, ok := session.Metadata["user_id"]
		if !ok {
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"received": true})
			return
		}

		amountCNY, _ := strconv.ParseInt(session.Metadata["amount_cny"], 10, 64)
		orderID := session.Metadata["order_id"]

		// 添加用户余额
		err = h.UserQuotaService.AddQuota(context.Background(), userID, amountCNY, "Stripe 充值", orderID)
		if err != nil {
			fmt.Printf("Failed to add quota for user %d: %v\n", userID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
