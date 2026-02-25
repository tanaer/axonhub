package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	googleConfig *oauth2.Config
	githubConfig *oauth2.Config
	jwtSecret    string
	frontendURL  string
}

// OAuthHandlerParams holds dependencies for OAuthHandler
type OAuthHandlerParams struct {
	fx.In
}

// NewOAuthHandler creates a new OAuth handler from environment variables
func NewOAuthHandler(params OAuthHandlerParams) *OAuthHandler {
	// Read config from environment variables
	googleClientID := os.Getenv("OAUTH_GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET")
	githubClientID := os.Getenv("OAUTH_GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("OAUTH_GITHUB_CLIENT_SECRET")
	jwtSecret := os.Getenv("OAUTH_JWT_SECRET")
	frontendURL := os.Getenv("OAUTH_FRONTEND_URL")

	if jwtSecret == "" {
		jwtSecret = "muskapi-oauth-default-secret-change-in-production"
	}
	if frontendURL == "" {
		frontendURL = "https://dev.claudeai.best"
	}

	h := &OAuthHandler{
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}

	// Configure Google OAuth
	if googleClientID != "" && googleClientSecret != "" {
		h.googleConfig = &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  frontendURL + "/auth/callback",
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	// Configure GitHub OAuth
	if githubClientID != "" && githubClientSecret != "" {
		h.githubConfig = &oauth2.Config{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  frontendURL + "/auth/callback",
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		}
	}

	return h
}

// OAuthClaims represents JWT claims for OAuth users
type OAuthClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Provider  string `json:"provider"`
	jwt.RegisteredClaims
}

// generateState generates a random state string for CSRF protection
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GoogleLogin initiates Google OAuth flow
func (h *OAuthHandler) GoogleLogin(c *gin.Context) {
	if h.googleConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google OAuth not configured"})
		return
	}
	state := generateState()
	// Store state in cookie for validation
	c.SetCookie("oauth_state", state, 300, "/", "", true, true)
	url := h.googleConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHubLogin initiates GitHub OAuth flow
func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	if h.githubConfig == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub OAuth not configured"})
		return
	}
	state := generateState()
	c.SetCookie("oauth_state", state, 300, "/", "", true, true)
	url := h.githubConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Callback handles OAuth callback from all providers
func (h *OAuthHandler) Callback(c *gin.Context) {
	// Validate state
	state := c.Query("state")
	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
		return
	}

	// Determine provider from referer or path
	var userInfo *UserInfo
	var provider string

	// Try Google first
	if h.googleConfig != nil {
		token, err := h.googleConfig.Exchange(context.Background(), code)
		if err == nil {
			userInfo, err = h.getGoogleUserInfo(token.AccessToken)
			if err == nil {
				provider = "google"
			}
		}
	}

	// Try GitHub if Google failed
	if userInfo == nil && h.githubConfig != nil {
		token, err := h.githubConfig.Exchange(context.Background(), code)
		if err == nil {
			userInfo, err = h.getGitHubUserInfo(token.AccessToken)
			if err == nil {
				provider = "github"
			}
		}
	}

	if userInfo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	// Generate JWT token
	claims := OAuthClaims{
		UserID:   userInfo.ID,
		Email:    userInfo.Email,
		Name:     userInfo.Name,
		Avatar:   userInfo.Avatar,
		Provider: provider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "MuskAPI",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set JWT cookie (secure=false for development, should be true in production with HTTPS)
	secure := h.frontendURL != "" && (len(h.frontendURL) > 5 && h.frontendURL[:5] == "https")
	c.SetCookie("auth_token", tokenString, 7*24*3600, "/", "", secure, true)

	// Redirect to frontend callback page
	redirectURL := h.frontendURL + "/auth/callback?auth=success"
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// UserInfo represents user information from OAuth provider
type UserInfo struct {
	ID     string
	Email  string
	Name   string
	Avatar string
}

// getGoogleUserInfo fetches user info from Google
func (h *OAuthHandler) getGoogleUserInfo(accessToken string) (*UserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &UserInfo{
		ID:     data.ID,
		Email:  data.Email,
		Name:   data.Name,
		Avatar: data.Picture,
	}, nil
}

// getGitHubUserInfo fetches user info from GitHub
func (h *OAuthHandler) getGitHubUserInfo(accessToken string) (*UserInfo, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	email := data.Email
	if email == "" {
		email = fmt.Sprintf("%s@users.noreply.github.com", data.Login)
	}

	return &UserInfo{
		ID:     fmt.Sprintf("%d", data.ID),
		Email:  email,
		Name:   data.Name,
		Avatar: data.AvatarURL,
	}, nil
}

// GetProviders returns list of available OAuth providers
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := []string{}
	if h.googleConfig != nil {
		providers = append(providers, "google")
	}
	if h.githubConfig != nil {
		providers = append(providers, "github")
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// VerifyToken verifies JWT token and returns user info
func (h *OAuthHandler) VerifyToken(c *gin.Context) {
	tokenString, err := c.Cookie("auth_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &OAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if claims, ok := token.Claims.(*OAuthClaims); ok && token.Valid {
		c.JSON(http.StatusOK, gin.H{
			"user_id":  claims.UserID,
			"email":    claims.Email,
			"name":     claims.Name,
			"avatar":   claims.Avatar,
			"provider": claims.Provider,
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
}

// Logout clears the auth cookie
func (h *OAuthHandler) Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// RequireAuth is a middleware that checks for valid JWT token
func (h *OAuthHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &OAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(h.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(*OAuthClaims); ok {
			c.Set("user", claims)
		}

		c.Next()
	}
}

// GetUserInfo extracts user info from context
func GetUserInfo(c *gin.Context) *OAuthClaims {
	if user, exists := c.Get("user"); exists {
		if claims, ok := user.(*OAuthClaims); ok {
			return claims
		}
	}
	return nil
}
