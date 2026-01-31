package donation

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, m middleware.AppMiddleware) {
	handler := &handler{
		service:    s,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	api := r.Group("/donations")
	api.GET("/", h.GetAllDonations)
}

func (h *handler) GetAllDonations(c *gin.Context) {
	c.JSON(200, "Donations")
}
