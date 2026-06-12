package social_program_transaction

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
	h := &handler{
		service:    s,
		middleware: m,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("social-programs/subscriptions/invoices/:id/pay", h.middleware.RequireRoles(enum.RoleOrangTuaAsuh), h.CreateSocialProgramTransaction)

	r.POST("/admin/social-programs/subscriptions/invoices/:id/pay-offline", h.middleware.RequireRoles(enum.RoleSocialManager), h.CreateOfflineSocialProgramTransaction)
}

// GetSocialProgramTransactionList
//
// @Summary List Social Program Transactions
// @Description Retrieve a paginated list of social program transactions (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/transactions [get]
func (h *handler) GetSocialProgramTransactionList(c *gin.Context) {
	ctx := c.Request.Context()

	var params SocialProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSocialProgramTransactionList(ctx, "", params)
	c.JSON(res.Status, res)
}

// GetSocialProgramTransactionByID
//
// @Summary Get Social Program Transaction by ID
// @Description Retrieve a specific social program transaction (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/transactions/{id} [get]
func (h *handler) GetSocialProgramTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetSocialProgramTransactionByID(ctx, id)
	c.JSON(res.Status, res)
}

// CreateSocialProgramTransaction
//
// @Summary Create Social Program Transaction
// @Description Initiate a Midtrans Snap payment for a social program invoice
// @Tags Social Programs
// @Accept json
// @Produce json
// @Param body body CreateTransactionRequest true "Transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/social-programs/transactions [post]
func (h *handler) CreateSocialProgramTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	accountID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateSocialProgramTransaction(ctx, accountID, id, req)
	c.JSON(res.Status, res)
}

// CreateOfflineSocialProgramTransaction
//
// @Summary Record Offline Social Program Transaction
// @Description Manually record an offline payment for a social program invoice (admin only)
// @Tags Social Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param body body CreateOfflineTransactionRequest true "Offline transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/admin/social-programs/transactions/pay-offline/{id} [post]
func (h *handler) CreateOfflineSocialProgramTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	invoiceID := c.Param("id")

	var req CreateOfflineTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", nil, nil))
		return
	}

	res := h.service.CreateOfflineSocialProgramTransaction(ctx, invoiceID, req)
	c.JSON(res.Status, res)
}

// GetMySocialProgramTransactionList
//
// @Summary List My Social Program Transactions
// @Description Retrieve a paginated list of social program transactions for the authenticated user
// @Tags Social Programs
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response
// @Router /api/me/social-programs/transactions [get]
func (h *handler) GetMySocialProgramTransactionList(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params SocialProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetMySocialProgramTransactionList(ctx, claims.AccountID, params)
	c.JSON(res.Status, res)
}

// GetMySocialProgramTransactionByID
//
// @Summary Get My Social Program Transaction by ID
// @Description Retrieve a specific social program transaction owned by the authenticated user
// @Tags Social Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/social-programs/transactions/{id} [get]
func (h *handler) GetMySocialProgramTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	res := h.service.GetMySocialProgramTransactionByID(ctx, id, claims.AccountID)
	c.JSON(res.Status, res)
}
