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
	public.POST("/:slug/transactions", h.middleware.AuthOptional(), h.CreateDonationProgramTransaction)
	public.GET("/:slug/transactions", h.GetPublicDonationProgramTransactionList)

	me := r.Group("/donation-programs/transactions/me")
	me.Use(h.middleware.RequireRoles(enum.RoleOrangTuaAsuh))
	{
		me.GET("", h.GetMyDonationProgramTransactionList)
		me.GET("/:id", h.GetMyDonationProgramTransactionByID)
	}

	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("/:id/transactions/monthly-income", h.GetDonationTransactionMonthlyIncome)
		admin.GET("/:id/transactions", h.GetDonationProgramTransactionList)
		admin.GET("/transactions/:id", h.GetDonationProgramTransactionByID)
		admin.POST("/:id/transactions", h.CreateOfflineDonationProgramTransaction)
		admin.POST("/transactions/:id/cancel", h.CancelOfflineDonationProgramTransaction)
		admin.GET("/:id/transactions/export", h.ExportDonationProgramTransactionCSV)
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
// @Param sortBy query string false "Sort order (e.g. gross_amount desc, created_at asc)"
// @Param startDate query string false "Filter start date (YYYY-MM-DD)"
// @Param endDate query string false "Filter end date (YYYY-MM-DD)"
// @Param search query string false "Search by donor name, email, or order ID"
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

// GetPublicDonationProgramTransactionList
//
// @Summary List Public Donation Program Transactions
// @Description Retrieve a paginated list of successful donation transactions for a specific donation program by slug
// @Tags Donation Programs
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Param sortBy query string false "Sort order (e.g. gross_amount desc, created_at asc)"
// @Param startDate query string false "Filter start date (YYYY-MM-DD)"
// @Param endDate query string false "Filter end date (YYYY-MM-DD)"
// @Param search query string false "Search by donor name, email, or order ID"
// @Success 200 {object} pkg.Response{data=DonationProgramTransactionListResponse}
// @Router /api/donation-programs/{slug}/transactions [get]
func (h *handler) GetPublicDonationProgramTransactionList(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	var params DonationProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetPublicDonationProgramTransactionList(ctx, slug, params)
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

func (h *handler) CancelOfflineDonationProgramTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	transactionID := c.Param("id")

	res := h.service.CancelOfflineDonationProgramTransaction(ctx, transactionID)
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
// @Param sortBy query string false "Sort order (e.g. gross_amount desc, created_at asc)"
// @Param startDate query string false "Filter start date (YYYY-MM-DD)"
// @Param endDate query string false "Filter end date (YYYY-MM-DD)"
// @Param search query string false "Search by donor name, email, or order ID"
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

// GetDonationTransactionMonthlyIncome
//
// @Summary Get Donation Program Monthly Income
// @Description Retrieve aggregated monthly incomes of a specific donation program for a given year (admin only)
// @Tags Donation Programs
// @Security BearerAuth
// @Produce json
// @Param id path string true "Donation Program ID"
// @Param year query string false "Filter by year"
// @Success 200 {object} pkg.Response{data=TransactionMonthlyIncomeRecord}
// @Router /api/admin/donation-programs/{id}/transactions/monthly-income [get]
func (h *handler) GetDonationTransactionMonthlyIncome(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var params MonthlyIncomeQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetDonationTransactionMonthlyIncome(ctx, id, params)
	c.JSON(res.Status, res)
}

// ExportDonationProgramTransactionCSV
//
// @Summary Export Donation Program Transaction as CSV
// @Description Export all transactions for a specific donation program as a CSV file (publicly accessible)
// @Tags Donation Programs
// @Produce text/csv
// @Param id path string true "Donation Program ID"
// @Param startDate query string false "Filter start date (YYYY-MM-DD, inclusive)"
// @Param endDate query string false "Filter end date (YYYY-MM-DD, inclusive)"
// @Success 200 {file} binary "CSV file"
// @Router /api/admin/donation-programs/{id}/transactions/export [get]
func (h *handler) ExportDonationProgramTransactionCSV(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	var params DonationProgramTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}

	csvBytes, filename, err := h.service.ExportDonationProgramTransactionCSV(ctx, donationProgramID, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)
}