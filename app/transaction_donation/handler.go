package transaction_donation

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
	// Public routes (anyone can donate or receive webhook)
	public := r.Group("/public/transaction-donations")
	public.POST("", h.CreateTransaction)
	public.POST("/notification", h.HandleNotification)

	// Admin-only routes
	protected := r.Group("/transaction-donations")
	protected.Use(h.middleware.RequireRoles(enum.RoleSuperadmin, enum.RoleFinance))
	{
		protected.GET("", h.ListTransactions)
		protected.GET("/:id", h.GetTransactionByID)
	}
}

// CreateTransaction
//
// @Summary Create Donation Transaction
// @Description Initiate a Midtrans Snap payment for a donation
// @Tags Transaction Donations
// @Accept json
// @Produce json
// @Param body body CreateTransactionRequest true "Transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/public/transaction-donations [post]
func (h *handler) CreateTransaction(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateTransaction(ctx, req)
	c.JSON(res.Status, res)
}

// HandleNotification
//
// @Summary Midtrans Payment Notification
// @Description Webhook endpoint for Midtrans to send payment status updates
// @Tags Transaction Donations
// @Accept json
// @Produce json
// @Param body body MidtransNotificationRequest true "Midtrans notification payload"
// @Success 200 {object} pkg.Response
// @Router /api/public/transaction-donations/notification [post]
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
// @Tags Transaction Donations
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param donation_id query string false "Filter by donation ID"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/transaction-donations [get]
func (h *handler) ListTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	var params QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.List(ctx, params)
	c.JSON(res.Status, res)
}

// GetTransactionByID
//
// @Summary Get Donation Transaction by ID
// @Description Retrieve a specific donation transaction (admin only)
// @Tags Transaction Donations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/transaction-donations/{id} [get]
func (h *handler) GetTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetByID(ctx, id)
	c.JSON(res.Status, res)
}
