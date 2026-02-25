package auth

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/token"
	"github.com/looplj/axonhub/conf"
)

// Service wraps the auth service
type Service struct {
	Authenticator *auth.Service
	config        conf.OAuthConfig
}

// NewService creates a new auth service
func NewService(config conf.OAuthConfig) *Service {
	// Get JWT secret from config or use default
	jwtSecret := config.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "muskapi-default-secret-change-in-production"
	}

	// Ensure URL doesn't have trailing slash
	baseURL := strings.TrimSuffix(config.URL, "/")

	options := auth.Opts{
		SecretReader: token.SecretFunc(func(id string) (string, error) {
			return jwtSecret, nil
		}),
		TokenDuration:  time.Hour * 24 * 7, // 7 days
		CookieDuration: time.Hour * 24 * 7, // 7 days
		Issuer:         "MuskAPI",
		URL:            baseURL,
		AvatarStore:    avatar.NewLocalFS("/tmp/muskapi-avatars"),
		DisableXSRF:    false,
		Logger:         log.New(os.Stderr, "[AUTH] ", log.LstdFlags),
	}

	service := auth.NewService(options)

	// Add Google provider if configured
	if config.GoogleClientID != "" && config.GoogleClientSecret != "" {
		service.AddProvider("google", config.GoogleClientID, config.GoogleClientSecret)
	}

	// Add GitHub provider if configured
	if config.GitHubClientID != "" && config.GitHubClientSecret != "" {
		service.AddProvider("github", config.GitHubClientID, config.GitHubClientSecret)
	}

	return &Service{
		Authenticator: service,
		config:        config,
	}
}

// Handlers returns auth and avatar handlers
func (s *Service) Handlers() (http.Handler, http.Handler) {
	return s.Authenticator.Handlers()
}

// Middleware returns auth middleware
func (s *Service) Middleware() *auth.Middleware {
	return s.Authenticator.Middleware()
}

// GetUser extracts user info from request
func GetUser(r *http.Request) (token.User, error) {
	return token.GetUserInfo(r)
}

// MustGetUser extracts user info from request (panics on error)
func MustGetUser(r *http.Request) token.User {
	return token.MustGetUserInfo(r)
}
