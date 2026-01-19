package auth

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
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
func (h *Handler) Register(c *gin.Context) {
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
func (h *Handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.Login(ctx, req)
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
func (h *Handler) ForgetPassword(c *gin.Context) {
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
func (h *Handler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}

	res := h.service.ResetPassword(ctx, req)
	c.JSON(res.Status, res)
}
