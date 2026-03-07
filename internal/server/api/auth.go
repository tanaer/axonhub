package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
)

type AuthHandlersParams struct {
	fx.In

	AuthService *biz.AuthService
	UserService *biz.UserService
}

func NewAuthHandlers(params AuthHandlersParams) *AuthHandlers {
	return &AuthHandlers{
		AuthService: params.AuthService,
		UserService: params.UserService,
	}
}

type AuthHandlers struct {
	AuthService *biz.AuthService
	UserService *biz.UserService
}

// SignInRequest 登录请求.
type SignInRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SignInResponse 登录响应.
type SignInResponse struct {
	User  *objects.UserInfo `json:"user"`
	Token string            `json:"token"`
}

// SignUpRequest 注册请求.
type SignUpRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

// SignUpResponse 注册响应.
type SignUpResponse struct {
	Message string            `json:"message"`
	Success bool              `json:"success"`
	User    *objects.UserInfo `json:"user"`
}

// ForgotPasswordRequest 忘记密码请求.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ForgotPasswordResponse 忘记密码响应.
type ForgotPasswordResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// SignIn handles user authentication.
func (h *AuthHandlers) SignIn(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req SignInRequest
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request format"))
		return
	}

	// Authenticate user
	user, err := h.AuthService.AuthenticateUser(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidPassword) {
			JSONError(c, http.StatusUnauthorized, errors.New("Invalid email or password"))
			return
		}

		JSONError(c, http.StatusInternalServerError, errors.New("Internal server error"))

		return
	}

	// Generate JWT token
	token, err := h.AuthService.GenerateJWTToken(ctx, user)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, errors.New("Internal server error"))
		return
	}

	response := SignInResponse{
		User:  biz.ConvertUserToUserInfo(ctx, user),
		Token: token,
	}

	c.JSON(http.StatusOK, response)
}

// SignUp handles user registration.
func (h *AuthHandlers) SignUp(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req SignUpRequest
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request format"))
		return
	}

	// Create user input
	input := ent.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
	}

	if req.Name != "" {
		input.FirstName = &req.Name
	}

	// Create user with system bypass (no auth required for signup)
	user, err := authz.RunWithSystemBypass(ctx, "auth-signup", func(bypassCtx context.Context) (*ent.User, error) {
		return h.UserService.CreateUser(bypassCtx, input)
	})
	if err != nil {
		// Check for duplicate email error
		JSONError(c, http.StatusBadRequest, errors.New("Email already registered or invalid data"))
		return
	}

	response := SignUpResponse{
		Message: "Account created successfully",
		Success: true,
		User:    biz.ConvertUserToUserInfo(ctx, user),
	}

	c.JSON(http.StatusOK, response)
}

// ForgotPassword handles password reset requests.
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	var (
		req ForgotPasswordRequest
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request format"))
		return
	}

	// TODO: Implement actual password reset email sending
	// For now, just return success to not leak whether email exists

	response := ForgotPasswordResponse{
		Message: "If the email exists, a reset link has been sent",
		Success: true,
	}

	c.JSON(http.StatusOK, response)
}
