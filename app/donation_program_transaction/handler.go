package donation_program_transaction

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
	public := r.Group("/donation-programs")
	public.POST("/:slug/transactions", h.CreateDonationProgramTransaction).Use(h.middleware.AuthOptional())

	me := r.Group("/me/donation-programs/transactions")
	me.Use(h.middleware.AuthRequired())
	{
		me.GET("", h.GetMyDonationProgramTransactionList)
		me.GET("/:id", h.GetMyDonationProgramTransactionByID)
	}

	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("/:id/transactions", h.GetDonationProgramTransactionList)
		admin.GET("/transactions/:id", h.GetDonationProgramTransactionByID)
		admin.POST("/:id/transactions", h.CreateOfflineDonationProgramTransaction)
	}
}

// GetDonationProgramTransactionList
//
// @Summary List Donation Programs Transactions
// @Description Retrieve a paginated list of donation transactions (admin only)
// @Tags Donation Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Filter by donation program ID"
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=DonationProgramTransactionListResponse}
// @Router /api/admin/donation-programs/{id}/transactions [get]
func (h *handler) GetDonationProgramTransactionList(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	var params DonationProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetDonationProgramTransactionList(ctx, "", donationProgramID, params)
	c.JSON(res.Status, res)
}

// GetDonationProgramTransactionByID
//
// @Summary Get Donation Program Transaction by ID
// @Description Retrieve a specific donation transaction (admin only)
// @Tags Donation Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/transactions/{id} [get]
func (h *handler) GetDonationProgramTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetDonationProgramTransactionByID(ctx, id)
	c.JSON(res.Status, res)
}

// CreateOfflineDonationProgramTransaction
//
// @Summary Create Offline Donation Program Transaction
// @Description Create a donation transaction without initiating a Midtrans payment (admin only)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Program ID"
// @Param body body CreateDonationProgramTransactionRequest true "Offline transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/transactions [post]
func (h *handler) CreateOfflineDonationProgramTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req CreateDonationProgramTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.CreateOfflineDonationProgramTransaction(ctx, claims.AccountID, donationProgramID, req)
	c.JSON(res.Status, res)
}

// CreateDonationProgramTransaction
//
// @Summary Create Donation Program Transaction
// @Description Initiate a Midtrans Snap payment for a donation
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Param body body CreateDonationProgramTransactionRequest true "Transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/donation-programs/{slug}/transactions [post]
func (h *handler) CreateDonationProgramTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	donationSlug := c.Param("slug")
	accountID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}
	var req CreateDonationProgramTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateDonationProgramTransaction(ctx, accountID, donationSlug, req)
	c.JSON(res.Status, res)
}

// GetMyDonationProgramTransactionList
//
// @Summary List My Donation Programs Transactions
// @Description Retrieve a paginated list of donation transactions for the authenticated user
// @Tags Donation Programs
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=DonationProgramTransactionListResponse}
// @Router /api/me/donation-programs/transactions [get]
func (h *handler) GetMyDonationProgramTransactionList(c *gin.Context) {
	ctx := c.Request.Context()

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params DonationProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetMyDonationProgramTransactionList(ctx, claims.AccountID, params)
	c.JSON(res.Status, res)
}

// GetMyDonationProgramTransactionByID
//
// @Summary Get My Donation Program Transaction by ID
// @Description Retrieve a specific donation transaction owned by the authenticated user
// @Tags Donation Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/donation-programs/transactions/{id} [get]
func (h *handler) GetMyDonationProgramTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	res := h.service.GetMyDonationProgramTransactionByID(ctx, id, claims.AccountID)
	c.JSON(res.Status, res)
}
