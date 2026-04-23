package donation_program_expense

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
	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET(":id/expenses", h.GetDonationProgramExpenseList)
		admin.GET("/expenses/:id", h.GetDonationProgramExpenseByID)
		admin.POST(":id/expenses", h.CreateDonationProgramExpense)
		admin.DELETE("expenses/:id", h.DeleteDonationProgramExpense)
	}
}

// GetDonationProgramExpenseList
//
// @Summary Get Donation Program Expense List
// @Description Get detailed information of all expenses (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Donation Program ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/expenses [get]
func (h *handler) GetDonationProgramExpenseList(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	var req DonationProgramExpenseQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.GetDonationProgramExpenseList(ctx, donationProgramID, req)
	c.JSON(resp.Status, resp)
}

// GetDonationProgramExpenseByID
//
// @Summary Get Donation Program Expense by ID
// @Description Get detailed information of a specific donation program expense entry (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/expenses/{id} [get]
func (h *handler) GetDonationProgramExpenseByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	resp := h.service.GetDonationProgramExpenseByID(ctx, id)
	c.JSON(resp.Status, resp)
}

// CreateDonationProgramExpense
//
// @Summary Create Donation Program Expense
// @Description Create a new donation program expense entry (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Donation Program ID"
// @Param payload formData DonationProgramExpenseRequest true "Donation Program Expense Data"
// @Param proof_file formData file false "Proof File"
// @Success 201 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/expenses [post]
func (h *handler) CreateDonationProgramExpense(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	var req DonationProgramExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.CreateDonationProgramExpense(ctx, donationProgramID, &req)
	c.JSON(resp.Status, resp)
}

// DeleteDonationProgramExpense
//
// @Summary Delete Donation Program Expense
// @Description Delete a donation program expense entry (requires authentication and proper role)
// @Tags Donation Programs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/donation-programs/expenses/{id} [delete]
func (h *handler) DeleteDonationProgramExpense(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	resp := h.service.DeleteDonationProgramExpense(ctx, id)
	c.JSON(resp.Status, resp)
}
