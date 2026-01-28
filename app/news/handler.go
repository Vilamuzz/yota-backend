package news

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(s Service, r *gin.RouterGroup, m middleware.AppMiddleware) {
	handler := &handler{
		service:    s,
		middleware: m,
	}
	handler.registerRoutes(r)
}

func (h *handler) registerRoutes(r *gin.RouterGroup) {
	api := r.Group("/news")

	protected := api.Group("")
	protected.Use(h.middleware.RequireRoles(string(user.RolePublicationManager)))
	{
		protected.GET("/", h.GetAllNews)
	}
}

// GetAllNews
// @Summary Get All News
// @Description Retrieve all news articles
// @Tags News
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/news/ [get]
func (h *handler) GetAllNews(c *gin.Context) {
	c.JSON(200, "News")
}
