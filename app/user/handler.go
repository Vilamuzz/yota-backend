package user

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/pkg"
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
	r.GET("/me", h.middleware.AuthRequired(), h.GetMe)
	r.PUT("/me", h.middleware.AuthRequired(), h.UpdateProfile)
	r.PUT("/me/password", h.middleware.AuthRequired(), h.UpdatePassword)
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

// GetMe
//
// @Summary Get Current User
// @Description Get details of the currently authenticated user
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response
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
	res := h.service.GetUserDetail(ctx, claims.UserID)
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
