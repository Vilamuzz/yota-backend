package admin_http

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
)

func (h *routeAdmin) handleAuthRoute(path string) {
	api := h.route.Group(path)
	api.POST("/login", h.Login)
}

// Login
//
// @Summary Login Admin
// @Description Login Admin
// @Tags Auth-Admin
// @Accept json
// @Produce json
// @Param payload body request.UserLoginRequest true "Login Admin"
// @Success 200 {object} pkg.Response
// @Router /admin/auth/login [post]
func (h *routeAdmin) Login(c *gin.Context) {
	ctx := c.Request.Context()
	req := request.UserLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.usecase.LoginAdmin(ctx, req)
	c.JSON(res.Status, res)
}
