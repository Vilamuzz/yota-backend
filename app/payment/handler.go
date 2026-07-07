package payment

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/foster_children_transaction"
	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/social_program_transaction"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/gin-gonic/gin"
)

type handler struct {
	donationService       donation_program_transaction.Service
	socialService         social_program_transaction.Service
	fosterChildrenService foster_children_transaction.Service
	paymentService        Service
	middleware            middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, ds donation_program_transaction.Service, ss social_program_transaction.Service, fs foster_children_transaction.Service, ps Service, m middleware.AppMiddleware) {
	h := &handler{
		donationService:       ds,
		socialService:         ss,
		fosterChildrenService: fs,
		paymentService:        ps,
		middleware:            m,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/webhooks/midtrans/notification", h.HandleMidtransNotification)

	adminGroup := r.Group("admin/payment-methods")
	adminGroup.Use(h.middleware.RequireRoles(enum.RoleSuperadmin))
	{
		adminGroup.GET("", h.GetPaymentMethods)
		adminGroup.PUT("/:id", h.UpdatePaymentMethod)
	}
}

// HandleMidtransNotification
//
// @Summary Unified Midtrans Payment Notification
// @Description Webhook endpoint for Midtrans to send payment status updates for all transaction types
// @Tags Payments
// @Accept json
// @Produce json
// @Param body body MidtransNotificationRequest true "Midtrans notification payload"
// @Success 200 {object} pkg.Response
// @Router /api/webhooks/midtrans/notification [post]
// MidtransNotificationRequest is an alias for swagger docs
type MidtransNotificationRequest payment_pkg.MidtransNotificationRequest
func (h *handler) HandleMidtransNotification(c *gin.Context) {
	ctx := c.Request.Context()

	var notification payment_pkg.MidtransNotificationRequest
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid notification payload", nil, nil))
		return
	}

	// Route based on OrderID prefix
	var res pkg.Response
	if strings.HasPrefix(notification.OrderID, "DON-") {
		res = h.donationService.HandleNotification(ctx, notification)
	} else if strings.HasPrefix(notification.OrderID, "SPI-") {
		res = h.socialService.HandleNotification(ctx, notification)
	} else if strings.HasPrefix(notification.OrderID, "FC-") {
		res = h.fosterChildrenService.HandleNotification(ctx, notification)
	} else {
		c.JSON(http.StatusOK, pkg.NewResponse(http.StatusOK, "Unknown order prefix, notification ignored", nil, nil))
		return
	}

	c.JSON(res.Status, res)
}

// GetPaymentMethods lists all payment methods
// @Summary Get payment methods
// @Description Get all payment methods configuration
// @Tags PaymentMethods
// @Security BearerAuth
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/admin/payment-methods [get]
func (h *handler) GetPaymentMethods(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.paymentService.GetPaymentMethods(ctx)
	c.JSON(res.Status, res)
}

// UpdatePaymentMethod updates payment method details
// @Summary Update payment method
// @Description Update payment method details like fee value, fee type, or active state
// @Tags PaymentMethods
// @Security BearerAuth
// @Param id path integer true "Payment Method ID"
// @Param body body UpdatePaymentMethodRequest true "Update payload"
// @Accept json
// @Produce json
// @Success 200 {object} pkg.Response
// @Failure 400 {object} pkg.Response
// @Router /api/admin/payment-methods/{id} [put]
func (h *handler) UpdatePaymentMethod(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Format ID tidak valid", nil, nil))
		return
	}

	var payload UpdatePaymentMethodRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Data payload tidak valid: "+err.Error(), nil, nil))
		return
	}

	res := h.paymentService.UpdatePaymentMethod(ctx, id, payload)
	c.JSON(res.Status, res)
}
