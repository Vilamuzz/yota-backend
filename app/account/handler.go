package account

import (
	"net/http"
	"strconv"

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
	me := r.Group("/me")
	me.Use(h.middleware.AuthRequired())
	{
		me.GET("", h.GetMe)
		me.PATCH("/profile", h.UpdateUserProfile)
		me.PATCH("/password", h.UpdatePassword)
		me.PATCH("/roles/default")
	}

	admin := r.Group("/admin/accounts")
	admin.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		admin.GET("/roles", h.GetRoleList)
		admin.GET("", h.GetAccountList)
		admin.GET("/:accountId", h.GetAccountByID)
		admin.PATCH("/:accountId/ban", h.BanAccount)
		admin.POST("/:accountId/roles/:roleId", h.AddAccountRole)
		admin.PATCH("/:accountId/roles/:roleId", h.UpdateAccountRole)
	}
}

// GetAccountList
//
// @Summary Get Accounts List
// @Description Get a list of accounts
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search query"
// @Param role_id query int false "Role ID filter"
// @Param is_banned query boolean false "Status filter"
// @Param sort_by query string false "Sort by"
// @Param sort_order query string false "Sort order"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=[]AccountResponse}
// @Router /api/admin/accounts [get]
func (h *handler) GetAccountList(c *gin.Context) {
	ctx := c.Request.Context()
	var queryParam AccountQueryParam
	if err := c.ShouldBindQuery(&queryParam); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.service.GetAccountList(ctx, queryParam)
	c.JSON(res.Status, res)
}

// GetAccountByID
//
// @Summary Get Account Detail
// @Description Get detailed information of an account by ID
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Success 200 {object} pkg.Response{data=AccountResponse}
// @Router /api/admin/accounts/{accountId} [get]
func (h *handler) GetAccountByID(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("accountId")
	res := h.service.GetAccountByID(ctx, accountID)
	c.JSON(res.Status, res)
}

// SetAccountBanStatus
//
// @Summary Set Account Ban Status
// @Description Set ban status of an account by ID
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param payload body SetAccountBanStatusRequest true "Ban status"
// @Success 200 {object} pkg.Response
// @Router /api/admin/accounts/{accountId}/ban [patch]
func (h *handler) BanAccount(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("accountId")
	var req SetAccountBanStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.service.SetAccountBanStatus(ctx, accountID, req)
	c.JSON(res.Status, res)
}

// AddAccountRole
//
// @Summary Add Account Role
// @Description Add a new role for an account
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param roleId path int true "Role ID"
// @Success 201 {object} pkg.Response
// @Router /api/admin/accounts/{accountId}/roles/{roleId} [post]
func (h *handler) AddAccountRole(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("accountId")
	roleID, _ := strconv.Atoi(c.Param("roleId"))
	res := h.service.AddAccountRole(ctx, accountID, roleID)
	c.JSON(res.Status, res)
}

// UpdateAccountRole
//
// @Summary Update Account Role
// @Description Update role status or default for an account
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param accountId path string true "Account ID"
// @Param roleId path int true "Role ID"
// @Param payload body UpdateAccountRoleRequest true "Update Role"
// @Success 200 {object} pkg.Response
// @Router /api/admin/accounts/{accountId}/roles/{roleId} [patch]
func (h *handler) UpdateAccountRole(c *gin.Context) {
	ctx := c.Request.Context()
	accountID := c.Param("accountId")
	roleID, _ := strconv.Atoi(c.Param("roleId"))
	req := UpdateAccountRoleRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.UpdateAccountRole(ctx, accountID, roleID, req)
	c.JSON(res.Status, res)
}

// GetMe
//
// @Summary Get Current User
// @Description Get details of the currently authenticated user
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response{data=UserProfileResponse}
// @Router /api/me [get]
func (h *handler) GetMe(c *gin.Context) {
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

	res := h.service.GetAccountByID(ctx, claims.AccountID)
	c.JSON(res.Status, res)
}

// UpdateUserProfile
//
// @Summary Update User Profile
// @Description Update profile information of the currently authenticated user
// @Tags Account
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param payload formData UpdateUserProfileRequest true "Update Profile"
// @Param profile_picture formData file false "Profile Picture"
// @Success 200 {object} pkg.Response
// @Router /api/me/profile [patch]
func (h *handler) UpdateUserProfile(c *gin.Context) {
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
	var req UpdateUserProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request", nil, nil))
		return
	}
	res := h.service.UpdateUserProfile(ctx, claims.AccountID, req)
	c.JSON(res.Status, res)
}

// UpdatePassword
//
// @Summary Update User Password
// @Description Update password of the currently authenticated user
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body UpdatePasswordRequest true "Update Password"
// @Success 200 {object} pkg.Response
// @Router /api/me/password [patch]
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
	res := h.service.UpdatePassword(ctx, claims.AccountID, req)
	c.JSON(res.Status, res)
}

// GetRoleList
//
// @Summary Get Role List
// @Description Get a list of roles
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response{data=RolesResponse}
// @Router /api/admin/accounts/roles [get]
func (h *handler) GetRoleList(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetRoleList(ctx)
	c.JSON(res.Status, res)
}
