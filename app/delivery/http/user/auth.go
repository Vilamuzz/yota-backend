package user_http

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
)

func (h *routeUser) handleAuthRoute(prefix string) {
	api := h.Route.Group(prefix)
	api.POST("/login", h.Login)
	api.POST("/register", h.Register)
	api.POST("/forget-password", h.ForgetPassword)
	api.POST("/reset-password", h.ResetPassword)
}

// Register
//
// @Summary Register User
// @Description Register User
// @Tags Auth-User
// @Accept json
// @Produce json
// @Param payload body request.RegisterRequest true "Register User"
// @Success 201 {object} pkg.Response
// @Router /api/user/auth/register [post]
func (h *routeUser) Register(c *gin.Context) {
	ctx := c.Request.Context()
	req := request.RegisterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.Usecase.RegisterUser(ctx, req)
	c.JSON(res.Status, res)
}

// Login
//
// @Summary Login User
// @Description Login User
// @Tags Auth-User
// @Accept json
// @Produce json
// @Param payload body request.LoginRequest true "Login User"
// @Success 200 {object} pkg.Response
// @Router /api/user/auth/login [post]
func (h *routeUser) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req := request.LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.Usecase.LoginUser(ctx, req)
	c.JSON(res.Status, res)
}

// ForgetPassword
//
// @Summary Forget Password
// @Description Send password reset email
// @Tags Auth-User
// @Accept json
// @Produce json
// @Param payload body request.ForgetPasswordRequest true "Forget Password"
// @Success 200 {object} pkg.Response
// @Router /api/user/auth/forget-password [post]
func (h *routeUser) ForgetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	req := request.ForgetPasswordRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.Usecase.ForgetPassword(ctx, req)
	c.JSON(res.Status, res)
}

// ResetPassword
//
// @Summary Reset Password
// @Description Reset password using token
// @Tags Auth-User
// @Accept json
// @Produce json
// @Param payload body request.ResetPasswordRequest true "Reset Password"
// @Success 200 {object} pkg.Response
// @Router /api/user/auth/reset-password [post]
func (h *routeUser) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	req := request.ResetPasswordRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.Usecase.ResetPassword(ctx, req)
	c.JSON(res.Status, res)
}
