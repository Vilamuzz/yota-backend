package user

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
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

	public := r.Group("/public/users")
	{
		public.GET("/roles", h.GetRoles)
	}

	me := r.Group("/me")
	me.Use(h.middleware.AuthRequired())
	{
		me.GET("", h.GetMe)
		me.PUT("", h.UpdateProfile)
		me.PUT("/password", h.UpdatePassword)
	}

	protected := r.Group("/users")
	protected.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		protected.GET("", h.GetUsersList)
		protected.GET("/:id", h.GetUserDetail)
		protected.PUT("/:id", h.UpdateUser)
	}
}

// GetRoles
//
// @Summary Get Roles
// @Description Get a list of roles
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response{data=[]RoleResponse}
// @Router /api/public/users/roles [get]
func (h *handler) GetRoles(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetRoles(ctx)
	c.JSON(res.Status, res)
}

// GetUsersList
//
// @Summary Get Users List
// @Description Get a list of users
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Param search query string false "Search query"
// @Param role query int false "Role ID filter"
// @Param status query boolean false "Status filter"
// @Success 200 {object} pkg.Response{data=[]UserResponse}
// @Router /api/users [get]
func (h *handler) GetUsersList(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParam UserQueryParam
	if err := c.ShouldBindQuery(&queryParam); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
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
// @Success 200 {object} pkg.Response{data=UserResponse}
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

// GetMe
//
// @Summary Get Current User
// @Description Get details of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response{data=UserProfileResponse}
// @Router /api/me [get]
func (h *handler) GetMe(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user claims from context
	userData, exists := c.Get("user_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "User not authenticated", nil, nil))
		return
	}

	// Type assert to UserJWTClaims
	claims, ok := userData.(jwt_pkg.UserJWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Invalid user data", nil, nil))
		return
	}

	// Get user details using the UserID from claims
	res := h.service.GetProfile(ctx, claims.UserID)
	c.JSON(res.Status, res)
}

// UpdateProfile
//
// @Summary Update User Profile
// @Description Update profile information of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body UpdateProfileRequest true "Update Profile"
// @Success 200 {object} pkg.Response
// @Router /api/me [put]
func (h *handler) UpdateProfile(c *gin.Context) {
	ctx := c.Request.Context()

	userData, exists := c.Get("user_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "User not authenticated", nil, nil))
		return
	}
	claims, ok := userData.(jwt_pkg.UserJWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Invalid user data", nil, nil))
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.service.UpdateProfile(ctx, claims.UserID, req)
	c.JSON(res.Status, res)
}

// UpdatePassword
//
// @Summary Update User Password
// @Description Update password of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body UpdatePasswordRequest true "Update Password"
// @Success 200 {object} pkg.Response
// @Router /api/me/password [put]
func (h *handler) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	userData, exists := c.Get("user_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, pkg.NewResponse(http.StatusUnauthorized, "User not authenticated", nil, nil))
		return
	}
	claims, ok := userData.(jwt_pkg.UserJWTClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, pkg.NewResponse(http.StatusInternalServerError, "Invalid user data", nil, nil))
		return
	}
	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.service.UpdatePassword(ctx, claims.UserID, req)
	c.JSON(res.Status, res)
}
