package superadmin_http

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
)

func (h *routeSuperadmin) handleAuthRoute(path string) {
	api := h.route.Group(path)
	api.POST("/login", h.Login)
}

// Login
//
// @Summary Login Superadmin
// @Description Login Superadmin
// @Tags Auth-Superadmin
// @Accept json
// @Produce json
// @Param payload body request.UserLoginRequest true "Login Superadmin"
// @Success 200 {object} pkg.Response
// @Router /superadmin/auth/login [post]
func (h *routeSuperadmin) Login(c *gin.Context) {
	ctx := c.Request.Context()
	req := request.UserLoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.usecase.LoginSuperadmin(ctx, req)
	c.JSON(res.Status, res)
}
