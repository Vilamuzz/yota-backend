package auth

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

type handler struct {
	service     Service
	userService user.Service
	middleware  middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, u user.Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:     s,
		userService: u,
		middleware:  m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	api := r.Group("/auth")

	// Apply strict rate limiting to auth endpoints
	authRateLimit := h.middleware.AuthRateLimitHandler()

	api.POST("/register", authRateLimit, h.Register)
	api.POST("/login", authRateLimit, h.Login)
	api.POST("/forget-password", h.middleware.CustomRateLimitHandler(5, 1*time.Minute), h.ForgetPassword)
	api.POST("/reset-password", authRateLimit, h.ResetPassword)
	api.POST("/verify-email", h.VerifyEmail)
	api.POST("/resend-verification", h.middleware.CustomRateLimitHandler(3, 1*time.Minute), h.ResendVerification)
	api.GET("/oauth/:provider", h.middleware.CustomRateLimitHandler(10, 1*time.Minute), h.OAuthLogin)
	api.GET("/oauth/:provider/callback", h.OAuthCallback)
}

// Register
//
// @Summary Register User
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body RegisterRequest true "Register User"
// @Success 201 {object} pkg.Response
// @Router /api/auth/register [post]
func (h *handler) Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.Register(ctx, req)
	c.JSON(res.Status, res)
}

// Login
//
// @Summary Login User
// @Description Login to user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body LoginRequest true "Login User"
// @Success 200 {object} pkg.Response
// @Router /api/auth/login [post]
func (h *handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.Login(ctx, req)
	c.JSON(res.Status, res)
}

// VerifyEmail
//
// @Summary Verify Email
// @Description Verify user email with token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body VerifyEmailRequest true "Verify Email"
// @Success 200 {object} pkg.Response
// @Router /api/auth/verify-email [post]
func (h *handler) VerifyEmail(c *gin.Context) {
	ctx := c.Request.Context()

	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.VerifyEmail(ctx, req.Token)
	c.JSON(res.Status, res)
}

// ResendVerification
//
// @Summary Resend Verification Email
// @Description Resend email verification link
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body ResendVerificationRequest true "Resend Verification"
// @Success 200 {object} pkg.Response
// @Router /api/auth/resend-verification [post]
func (h *handler) ResendVerification(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.ResendVerificationEmail(ctx, req.Email)
	c.JSON(res.Status, res)
}

// ForgetPassword
//
// @Summary Forget Password
// @Description Send password reset email
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body ForgetPasswordRequest true "Forget Password"
// @Success 200 {object} pkg.Response
// @Router /api/auth/forget-password [post]
func (h *handler) ForgetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req ForgetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.ForgetPassword(ctx, req)
	c.JSON(res.Status, res)
}

// ResetPassword
//
// @Summary Reset Password
// @Description Reset password using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body ResetPasswordRequest true "Reset Password"
// @Success 200 {object} pkg.Response
// @Router /api/auth/reset-password [post]
func (h *handler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.ResetPassword(ctx, req)
	c.JSON(res.Status, res)
}

// OAuthLogin
//
// @Summary OAuth Login
// @Description Initiate OAuth login with Provider
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth Provider"
// @Router /api/auth/oauth/{provider} [get]
func (h *handler) OAuthLogin(c *gin.Context) {
	provider := c.Param("provider")
	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// OAuthCallback
//
// @Summary OAuth Callback
// @Description Handle OAuth callback from Provider
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth Provider"
// @Success 200 {object} pkg.Response
// @Router /api/auth/oauth/{provider}/callback [get]
func (h *handler) OAuthCallback(c *gin.Context) {
	ctx := c.Request.Context()
	provider := c.Param("provider")

	q := c.Request.URL.Query()
	q.Add("provider", provider)
	c.Request.URL.RawQuery = q.Encode()

	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "OAuth authentication failed", nil, nil))
		return
	}

	res := h.service.OAuthLogin(ctx, provider, gothUser)

	if res.Status == http.StatusOK {
		authRes := res.Data.(AuthResponse)
		frontendURL := os.Getenv("FE_URL")
		c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, authRes.Token))
		return
	}

	c.JSON(res.Status, res)
}
