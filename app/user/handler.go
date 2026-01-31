package user

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(s Service, m middleware.AppMiddleware) *handler {
	return &handler{
		service:    s,
		middleware: m,
	}
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	api := r.Group("/users")
	api.GET("", h.middleware.RequireRoles(string(RoleSuperadmin)), h.GetUsersList)
}

func (h *handler) GetUsersList(c *gin.Context) {
	ctx := c.Request.Context()
	queryParam := c.Request.URL.Query()
	res := h.service.GetUsersList(ctx, queryParam)
	c.JSON(res.Status, res)
}
