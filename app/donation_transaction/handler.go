package donation_transaction

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
	public := r.Group("/public/donation-transactions")
	public.Use(h.middleware.AuthRequired())
	{
		public.GET("/me", h.ListMyTransactions)
		public.GET("/me/:id", h.GetMyTransactionByID)
	}
	public.POST("", h.CreateTransaction).Use(h.middleware.AuthOptional())
	public.POST("/notification", h.HandleNotification)

	protected := r.Group("/donation-transactions")
	protected.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		protected.GET("", h.ListTransactions)
		protected.GET("/:id", h.GetTransactionByID)
		protected.POST("", h.CreateOfflineTransaction)
	}
}

// CreateOfflineTransaction
//
// @Summary Create Offline Donation Transaction
// @Description Create a donation transaction without initiating a Midtrans payment (admin only)
// @Tags Donation Transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateTransactionRequest true "Offline transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/donation-transactions [post]
func (h *handler) CreateOfflineTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.CreateOfflineTransaction(ctx, req, claims.UserID)
	c.JSON(res.Status, res)
}

// CreateTransaction
//
// @Summary Create Donation Transaction
// @Description Initiate a Midtrans Snap payment for a donation
// @Tags Donation Transactions
// @Accept json
// @Produce json
// @Param body body CreateTransactionRequest true "Transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/public/donation-transactions [post]
func (h *handler) CreateTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	userID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			userID = claims.UserID
		}
	}
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateTransaction(ctx, req, userID)
	c.JSON(res.Status, res)
}

// HandleNotification
//
// @Summary Midtrans Payment Notification
// @Description Webhook endpoint for Midtrans to send payment status updates
// @Tags Donation Transactions
// @Accept json
// @Produce json
// @Param body body MidtransNotificationRequest true "Midtrans notification payload"
// @Success 200 {object} pkg.Response
// @Router /api/public/donation-transactions/notification [post]
func (h *handler) HandleNotification(c *gin.Context) {
	ctx := c.Request.Context()

	var notification MidtransNotificationRequest
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid notification payload", nil, nil))
		return
	}

	res := h.service.HandleNotification(ctx, notification)
	c.JSON(res.Status, res)
}

// ListTransactions
//
// @Summary List Donation Transactions
// @Description Retrieve a paginated list of donation transactions (admin only)
// @Tags Donation Transactions
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param donation_id query string false "Filter by donation ID"
// @Param user_id query string false "Filter by user ID"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=DonationTransactionListResponse}
// @Router /api/donation-transactions [get]
func (h *handler) ListTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	var params DonationTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.ListTransactions(ctx, params)
	c.JSON(res.Status, res)
}

// ListMyTransactions
//
// @Summary List My Donation Transactions
// @Description Retrieve a paginated list of donation transactions for the authenticated user
// @Tags Donation Transactions
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param donation_id query string false "Filter by donation ID"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=DonationTransactionListResponse}
// @Router /api/public/donation-transactions/me [get]
func (h *handler) ListMyTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params DonationTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.ListMyTransactions(ctx, params, claims.UserID)
	c.JSON(res.Status, res)
}

// GetTransactionByID
//
// @Summary Get Donation Transaction by ID
// @Description Retrieve a specific donation transaction (admin only)
// @Tags Donation Transactions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/donation-transactions/{id} [get]
func (h *handler) GetTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetTransactionByID(ctx, id)
	c.JSON(res.Status, res)
}

// GetMyTransactionByID
//
// @Summary Get My Donation Transaction by ID
// @Description Retrieve a specific donation transaction owned by the authenticated user
// @Tags Donation Transactions
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/public/donation-transactions/me/{id} [get]
func (h *handler) GetMyTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	res := h.service.GetMyTransactionByID(ctx, id, claims.UserID)
	c.JSON(res.Status, res)
}
