package payment

import (
	"net/http"
	"strings"

	"github.com/Vilamuzz/yota-backend/app/donation_program_transaction"
	"github.com/Vilamuzz/yota-backend/app/foster_children_transaction"
	"github.com/Vilamuzz/yota-backend/app/social_program_transaction"
	"github.com/Vilamuzz/yota-backend/pkg"
	payment_pkg "github.com/Vilamuzz/yota-backend/pkg/payment"
	"github.com/gin-gonic/gin"
)

type handler struct {
	donationService       donation_program_transaction.Service
	socialService         social_program_transaction.Service
	fosterChildrenService foster_children_transaction.Service
}

func NewHandler(r *gin.RouterGroup, ds donation_program_transaction.Service, ss social_program_transaction.Service, fs foster_children_transaction.Service) {
	h := &handler{
		donationService:       ds,
		socialService:         ss,
		fosterChildrenService: fs,
	}
	h.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/midtrans/notification", h.HandleMidtransNotification)
}

// HandleMidtransNotification
//
// @Summary Unified Midtrans Payment Notification
// @Description Webhook endpoint for Midtrans to send payment status updates for all transaction types
// @Tags Payments
// @Accept json
// @Produce json
// @Param body body payment_pkg.MidtransNotificationRequest true "Midtrans notification payload"
// @Success 200 {object} pkg.Response
// @Router /api/webhooks/midtrans/notification [post]
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
		c.JSON(http.StatusNotFound, pkg.NewResponse(http.StatusNotFound, "Unknown order prefix", nil, nil))
		return
	}

	c.JSON(res.Status, res)
}
