package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/user"
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
	Email     string `json:"email"     binding:"required,email"`
	Password  string `json:"password"  binding:"required,min=7"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// ForgotPasswordRequest 忘记密码请求.
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
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

	// Create user with system bypass (public registration)
	firstName := req.FirstName
	lastName := req.LastName
	status := user.StatusActivated

	createInput := ent.CreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Status:   &status,
	}
	if firstName != "" {
		createInput.FirstName = &firstName
	}
	if lastName != "" {
		createInput.LastName = &lastName
	}

	newUser, err := authz.RunWithSystemBypass(ctx, "signup", func(bypassCtx context.Context) (*ent.User, error) {
		return h.UserService.CreateUser(bypassCtx, createInput)
	})
	
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Failed to create user: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account created successfully",
		"user":    biz.ConvertUserToUserInfo(ctx, newUser),
	})
}

// ForgotPassword handles password reset request.
func (h *AuthHandlers) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, errors.New("Invalid request format"))
		return
	}

	// TODO: Implement actual password reset email sending
	// For now, just return success to avoid revealing if email exists
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a reset link has been sent",
	})
}
