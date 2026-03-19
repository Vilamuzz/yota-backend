package donation_expense

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
	protected := r.Group("/donation-expenses")
	protected.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		protected.POST("/", h.CreateExpense)
		protected.PUT("/:id", h.UpdateExpense)
		protected.DELETE("/:id", h.DeleteExpense)
		protected.GET("/:id", h.GetExpenseByID)
		protected.GET("/", h.ListExpenses)
	}
}

// CreateExpense
//
// @Summary Create Expense
// @Description Create a new expense entry (requires authentication and proper role)
// @Tags Donation Expenses
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param donation_id formData string true "Donation ID"
// @Param amount formData number true "Expense Amount"
// @Param description formData string true "Expense Description"
// @Param image formData file false "Expense Image File"
// @Success 201 {object} pkg.Response
// @Router /api/donation-expenses/ [post]
func (h *handler) CreateExpense(c *gin.Context) {
	var req CreateExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.CreateExpense(c.Request.Context(), &req)
	c.JSON(resp.Status, resp)
}

// UpdateExpense
//
// @Summary Update Expense
// @Description Update an existing expense entry (requires authentication and proper role)
// @Tags Donation Expenses
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Expense ID"
// @Param donation_id formData string false "Donation ID"
// @Param amount formData number false "Expense Amount"
// @Param description formData string false "Expense Description"
// @Param image formData file false "Expense Image File"
// @Success 200 {object} pkg.Response
// @Router /api/donation-expenses/{id} [put]
func (h *handler) UpdateExpense(c *gin.Context) {
	var req UpdateExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	req.ID = c.Param("id")
	resp := h.service.UpdateExpense(c.Request.Context(), &req)
	c.JSON(resp.Status, resp)
}

// DeleteExpense
//
// @Summary Delete Expense
// @Description Delete an existing expense entry (requires authentication and proper role)
// @Tags Donation Expenses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/donation-expenses/{id} [delete]
func (h *handler) DeleteExpense(c *gin.Context) {
	resp := h.service.DeleteExpense(c.Request.Context(), c.Param("id"))
	c.JSON(resp.Status, resp)
}

// GetExpenseByID
//
// @Summary Get Expense by ID
// @Description Get detailed information of a specific expense entry (requires authentication and proper role)
// @Tags Donation Expenses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/donation-expenses/{id} [get]
func (h *handler) GetExpenseByID(c *gin.Context) {
	resp := h.service.GetExpenseByID(c.Request.Context(), c.Param("id"))
	c.JSON(resp.Status, resp)
}

// ListExpenses
//
// @Summary List Expenses
// @Description Get detailed information of all expenses (requires authentication and proper role)
// @Tags Donation Expenses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param donation_id query string false "Filter by donation ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/donation-expenses/ [get]
func (h *handler) ListExpenses(c *gin.Context) {
	var req DonationExpenseQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.ListExpenses(c.Request.Context(), req)
	c.JSON(resp.Status, resp)
}
