package prayer

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service Service
}

func NewHandler(service Service) handler {
	return handler{service: service}
}

func (h *handler) RegisterRoutes(router *gin.Engine) {
	prayerRoutes := router.Group("/prayers")
	prayerRoutes.GET("/:id", h.FindPrayerByID)
	prayerRoutes.GET("/", h.ListPrayers)
}

// @Summary Get Prayer by ID
// @Description Get a prayer by its ID
// @Tags Prayer
// @Accept json
// @Produce json
// @Param id path string true "Prayer ID"
// @Success 200 {object} pkg.Response{data=PrayerResponse}
// @Failure 400 {object} pkg.Response
// @Failure 404 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /prayers/{id} [get]
func (h *handler) FindPrayerByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.FindPrayerByID(ctx, id)
	c.JSON(res.Status, res)
}

// @Summary List Prayers
// @Description Get a list of prayers
// @Tags Prayer
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param search query string false "Search query"
// @Success 200 {object} pkg.Response{data=PrayerListResponse}
// @Failure 400 {object} pkg.Response
// @Failure 500 {object} pkg.Response
// @Router /prayers [get]
func (h *handler) ListPrayers(c *gin.Context) {
	ctx := c.Request.Context()
	var params PrayerQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}
	res := h.service.ListPrayers(ctx, params)
	c.JSON(res.Status, res)
}
