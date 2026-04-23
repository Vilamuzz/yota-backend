package foster_children_expense

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
	admin := r.Group("/admin/foster-children")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET(":id/expenses", h.GetFosterChildrenExpenseList)
		admin.GET("/expenses/:id", h.GetFosterChildrenExpenseByID)
		admin.POST(":id/expenses", h.CreateFosterChildrenExpense)
		admin.DELETE("expenses/:id", h.DeleteFosterChildrenExpense)
	}
}

// GetFosterChildrenExpenseList
//
// @Summary Get Foster Children Expense List
// @Description Get detailed information of all expenses for a foster child (requires authentication and proper role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Foster Children ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Items per page"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/{id}/expenses [get]
func (h *handler) GetFosterChildrenExpenseList(c *gin.Context) {
	ctx := c.Request.Context()
	fosterChildrenID := c.Param("id")

	var req FosterChildrenExpenseQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.GetFosterChildrenExpenseList(ctx, fosterChildrenID, req)
	c.JSON(resp.Status, resp)
}

// GetFosterChildrenExpenseByID
//
// @Summary Get Foster Children Expense by ID
// @Description Get detailed information of a specific foster children expense entry (requires authentication and proper role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/expenses/{id} [get]
func (h *handler) GetFosterChildrenExpenseByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	resp := h.service.GetFosterChildrenExpenseByID(ctx, id)
	c.JSON(resp.Status, resp)
}

// CreateFosterChildrenExpense
//
// @Summary Create Foster Children Expense
// @Description Create a new foster children expense entry (requires authentication and proper role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Foster Children ID"
// @Param payload formData FosterChildrenExpenseRequest true "Foster Children Expense Data"
// @Param proof_file formData file false "Proof File"
// @Success 201 {object} pkg.Response
// @Router /api/admin/foster-children/{id}/expenses [post]
func (h *handler) CreateFosterChildrenExpense(c *gin.Context) {
	ctx := c.Request.Context()
	fosterChildrenID := c.Param("id")

	var req FosterChildrenExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, err.Error(), nil, nil))
		return
	}
	resp := h.service.CreateFosterChildrenExpense(ctx, fosterChildrenID, &req)
	c.JSON(resp.Status, resp)
}

// DeleteFosterChildrenExpense
//
// @Summary Delete Foster Children Expense
// @Description Delete a foster children expense entry (requires authentication and proper role)
// @Tags Foster Children
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/expenses/{id} [delete]
func (h *handler) DeleteFosterChildrenExpense(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")
	resp := h.service.DeleteFosterChildrenExpense(ctx, id)
	c.JSON(resp.Status, resp)
}
