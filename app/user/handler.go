package user

import (
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
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
	api := r.Group("/users")
	api.GET("", h.middleware.RequireRoles(string(RoleSuperadmin)), h.GetUsersList)
	api.GET("/:id", h.middleware.RequireRoles(string(RoleSuperadmin)), h.GetUserDetail)
	api.PUT("/:id", h.middleware.RequireRoles(string(RoleSuperadmin)), h.UpdateUser)
}

// GetUsersList
//
// @Summary Get Users List
// @Description Get a list of users
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/users [get]
func (h *handler) GetUsersList(c *gin.Context) {
	ctx := c.Request.Context()
	queryParam := c.Request.URL.Query()
	res := h.service.GetUsersList(ctx, queryParam)
	c.JSON(res.Status, res)
}

// GetUserDetail
//
// @Summary Get User Detail
// @Description Get detailed information of a user by ID
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} pkg.Response
// @Router /api/users/{id} [get]
func (h *handler) GetUserDetail(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	res := h.service.GetUserDetail(ctx, userID)
	c.JSON(res.Status, res)
}

// UpdateUser
//
// @Summary Update User
// @Description Update user information by ID
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body UpdateUserRequest true "Update User Data"
// @Success 200 {object} pkg.Response
// @Router /api/users/{id} [put]
func (h *handler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")
	var updateData UpdateUserRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(400, pkg.NewResponse(400, "Invalid request body", nil, nil))
		return
	}
	res := h.service.UpdateUser(ctx, userID, updateData)
	c.JSON(res.Status, res)
}
