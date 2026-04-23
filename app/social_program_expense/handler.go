package social_program_expense

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
	// Admin routes
	admin := r.Group("/admin/social-programs")
	admin.Use(h.middleware.RequireRoles(enum.RoleSocialManager, enum.RoleFinance))
	{
		admin.GET("/:id/expenses", h.GetSocialProgramExpenseList)
		admin.GET("/expenses/:id", h.GetSocialProgramExpenseByID)
		admin.POST("/:id/expenses", h.CreateSocialProgramExpense)
		admin.DELETE("/expenses/:id", h.DeleteSocialProgramExpense)
	}
}

// GetSocialProgramExpenseList
//
// @Summary List Social Program Expenses
// @Description Retrieve a paginated list of expenses for a specific social program
// @Tags Social Program Expenses
// @Accept json
// @Produce json
// @Param id path string true "Social Program ID"
// @Param limit query int false "Pagination limit"
// @Param next_cursor query string false "Pagination cursor (next page)"
// @Param prev_cursor query string false "Pagination cursor (prev page)"
// @Success 200 {object} pkg.Response{data=SocialProgramExpenseListResponse}
// @Router /api/social-programs/{id}/expenses [get]
func (h *handler) GetSocialProgramExpenseList(c *gin.Context) {
	ctx := c.Request.Context()
	socialProgramID := c.Param("id")

	var queryParams SocialProgramExpenseQueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetSocialProgramExpenseList(ctx, socialProgramID, queryParams)
	c.JSON(res.Status, res)
}

// GetSocialProgramExpenseByID
//
// @Summary Get Social Program Expense by ID
// @Description Get detailed information of a specific social program expense
// @Tags Social Program Expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response{data=SocialProgramExpenseDetailResponse}
// @Router /api/social-programs/expenses/{id} [get]
func (h *handler) GetSocialProgramExpenseByID(c *gin.Context) {
	ctx := c.Request.Context()
	expenseID := c.Param("id")

	res := h.service.GetSocialProgramExpenseByID(ctx, expenseID)
	c.JSON(res.Status, res)
}

// CreateSocialProgramExpense
//
// @Summary Create Social Program Expense
// @Description Create a new expense for a social program (requires authentication and admin role)
// @Tags Social Program Expenses
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Social Program ID"
// @Param payload formData SocialProgramExpenseRequest true "Expense Data"
// @Param proof_file formData file false "Proof File"
// @Success 201 {object} pkg.Response
// @Router /api/admin/social-programs/{id}/expenses [post]
func (h *handler) CreateSocialProgramExpense(c *gin.Context) {
	ctx := c.Request.Context()
	socialProgramID := c.Param("id")

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req SocialProgramExpenseRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateSocialProgramExpense(ctx, claims.AccountID, socialProgramID, &req)
	c.JSON(res.Status, res)
}

// DeleteSocialProgramExpense
//
// @Summary Delete Social Program Expense
// @Description Delete a social program expense (requires authentication and admin role)
// @Tags Social Program Expenses
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Expense ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/social-programs/expenses/{id} [delete]
func (h *handler) DeleteSocialProgramExpense(c *gin.Context) {
	ctx := c.Request.Context()
	expenseID := c.Param("id")

	res := h.service.DeleteSocialProgramExpense(ctx, expenseID)
	c.JSON(res.Status, res)
}
