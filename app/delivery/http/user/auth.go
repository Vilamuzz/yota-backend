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
}

// Register
//
// @Summary Register User
// @Description Register User
// @Tags Auth-User
// @Accept json
// @Produce json
// @Param payload body request.UserRegisterRequest true "Register User"
// @Success 201 {object} pkg.Response
// @Router /user/auth/register [post]
func (h *routeUser) Register(c *gin.Context) {
	ctx := c.Request.Context()
	req := request.UserRegisterRequest{}
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
// @Param payload body request.UserLoginRequest true "Login User"
// @Success 200 {object} pkg.Response
// @Router /user/auth/login [post]
func (h *routeUser) Login(c *gin.Context) {
	ctx := c.Request.Context()

	req := request.UserLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.Usecase.LoginUser(ctx, req)
	c.JSON(res.Status, res)
}
