package donation_program_expense

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
	public.GET("/:slug/expenses", h.GetPublicDonationProgramExpenseList)
	public.GET("/expenses/:id", h.GetDonationProgramExpenseByID)

	admin := r.Group("/admin/donation-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("/:id/expenses", h.GetDonationProgramExpenseList)
		admin.GET("/expenses/:id", h.GetDonationProgramExpenseByID)
		admin.POST("/:id/expenses", h.CreateDonationProgramExpense)
		admin.DELETE("/expenses/:id", h.DeleteDonationProgramExpense)
	}
}

// GetPublicDonationProgramExpenseList
//
// @Summary Get Public Donation Program Expense List
// @Description Get paginated list of expenses for a specific donation program (publicly accessible)
// @Tags Donation Programs
// @Accept json
// @Produce json
// @Param slug path string true "Donation Program Slug"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/donation-programs/{slug}/expenses [get]
func (h *handler) GetPublicDonationProgramExpenseList(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	var req DonationProgramExpenseQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.GetPublicDonationProgramExpenseList(ctx, slug, req)
	c.JSON(resp.Status, resp)
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
// @Param title formData string true "Expense Title"
// @Param amount formData number true "Expense Amount"
// @Param expenseDate formData string true "Expense Date (YYYY-MM-DD)"
// @Param note formData string false "Expense Note"
// @Param proofFile formData file false "Proof File"
// @Success 201 {object} pkg.Response
// @Router /api/admin/donation-programs/{id}/expenses [post]
func (h *handler) CreateDonationProgramExpense(c *gin.Context) {
	ctx := c.Request.Context()
	donationProgramID := c.Param("id")

	var req DonationProgramExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body: "+err.Error(), nil, nil))
		return
	}
	userData, _ := c.Get("user_data")
	claims := userData.(jwt_pkg.UserJWTClaims)

	resp := h.service.CreateDonationProgramExpense(ctx, claims.AccountID, donationProgramID, &req)
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
	userData, _ := c.Get("user_data")
	claims := userData.(jwt_pkg.UserJWTClaims)

	resp := h.service.DeleteDonationProgramExpense(ctx, claims.AccountID, id)
	c.JSON(resp.Status, resp)
}
