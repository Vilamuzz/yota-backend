package foundation_profile

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/gin-gonic/gin"
)

type handler struct {
	service    Service
	middleware middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, m middleware.AppMiddleware) {
	h := &handler{
		service:    s,
		middleware: m,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	public := r.Group("/foundation-profile")
	public.GET("", h.GetFoundationProfile)

	admin := r.Group("/admin/foundation-profile")
	admin.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		admin.POST("", h.CreateFoundationProfile)
		admin.PUT("/:id", h.UpdateFoundationProfile)
	}
}

// GetFoundationProfile
//
// @Summary Get Foundation Profile
// @Description Retrieve the foundation profile information
// @Tags Foundation Profile
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response{data=FoundationProfileResponse}
// @Router /api/foundation-profile [get]
func (h *handler) GetFoundationProfile(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetFoundationProfile(ctx)
	c.JSON(res.Status, res)
}

// CreateFoundationProfile
//
// @Summary Create Foundation Profile
// @Description Create the foundation profile (requires chairman or superadmin role)
// @Tags Foundation Profile
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param foundationName formData string true "Foundation Name"
// @Param founderName formData string false "Founder Name"
// @Param founderPicture formData file false "Founder Picture"
// @Param foundationAddress formData string false "Foundation Address"
// @Param foundationPhone formData string false "Foundation Phone"
// @Param foundationEmail formData string false "Foundation Email"
// @Param foundationInstagram formData string false "Foundation Instagram"
// @Param foundationFacebook formData string false "Foundation Facebook"
// @Param foundationTwitter formData string false "Foundation Twitter"
// @Param embeddedAddress formData string false "Embedded Address"
// @Param logo formData file false "Foundation Logo"
// @Param icon formData file false "Foundation Icon"
// @Param organization_structure formData file false "Organization Structure Image"
// @Param hero_image_one formData file false "Hero Image 1"
// @Param hero_image_two formData file false "Hero Image 2"
// @Param hero_image_three formData file false "Hero Image 3"
// @Param hero_image_four formData file false "Hero Image 4"
// @Success 201 {object} pkg.Response{data=FoundationProfileResponse}
// @Router /api/admin/foundation-profile [post]
func (h *handler) CreateFoundationProfile(c *gin.Context) {
	ctx := c.Request.Context()

	var req FoundationProfileCreateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.CreateFoundationProfile(ctx, req)
	c.JSON(res.Status, res)
}

// UpdateFoundationProfile
//
// @Summary Update Foundation Profile
// @Description Update the foundation profile (requires chairman or superadmin role)
// @Tags Foundation Profile
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Foundation Profile ID"
// @Param foundationName formData string false "Foundation Name"
// @Param founderName formData string false "Founder Name"
// @Param founderPicture formData file false "Founder Picture"
// @Param foundationAddress formData string false "Foundation Address"
// @Param foundationPhone formData string false "Foundation Phone"
// @Param foundationEmail formData string false "Foundation Email"
// @Param foundationInstagram formData string false "Foundation Instagram"
// @Param foundationFacebook formData string false "Foundation Facebook"
// @Param foundationTwitter formData string false "Foundation Twitter"
// @Param embeddedAddress formData string false "Embedded Address"
// @Param logo formData file false "Foundation Logo"
// @Param icon formData file false "Foundation Icon"
// @Param organization_structure formData file false "Organization Structure Image"
// @Param hero_image_one formData file false "Hero Image 1"
// @Param hero_image_two formData file false "Hero Image 2"
// @Param hero_image_three formData file false "Hero Image 3"
// @Param hero_image_four formData file false "Hero Image 4"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foundation-profile/{id} [put]
func (h *handler) UpdateFoundationProfile(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req FoundationProfileUpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}

	res := h.service.UpdateFoundationProfile(ctx, id, req)
	c.JSON(res.Status, res)
}
