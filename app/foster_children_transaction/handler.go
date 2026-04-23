package foster_children_transaction

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
	public := r.Group("/foster-children")
	public.POST("/:id/transactions", h.CreateFosterChildrenTransaction).Use(h.middleware.AuthOptional())

	me := r.Group("/me/foster-children/transactions")
	me.Use(h.middleware.AuthRequired())
	{
		me.GET("", h.GetMyFosterChildrenTransactionList)
		me.GET("/:id", h.GetMyFosterChildrenTransactionByID)
	}

	admin := r.Group("/admin/foster-children")
	admin.Use(h.middleware.RequireRoles(enum.RoleFinance))
	{
		admin.GET("/:id/transactions", h.GetFosterChildrenTransactionList)
		admin.GET("/transactions/:id", h.GetFosterChildrenTransactionByID)
		admin.POST("/:id/transactions", h.CreateOfflineFosterChildrenTransaction)
	}
}

// GetFosterChildrenTransactionList
//
// @Summary List Foster Children Transactions
// @Description Retrieve a paginated list of foster children transactions (admin only)
// @Tags Foster Children
// @Security BearerAuth
// @Produce json
// @Param id path string true "Filter by foster children ID"
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=FosterChildrenTransactionListResponse}
// @Router /api/admin/foster-children/{id}/transactions [get]
func (h *handler) GetFosterChildrenTransactionList(c *gin.Context) {
	ctx := c.Request.Context()
	fosterChildrenID := c.Param("id")

	var params FosterChildrenTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetFosterChildrenTransactionList(ctx, "", fosterChildrenID, params)
	c.JSON(res.Status, res)
}

// GetFosterChildrenTransactionByID
//
// @Summary Get Foster Children Transaction by ID
// @Description Retrieve a specific foster children transaction (admin only)
// @Tags Foster Children
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/admin/foster-children/transactions/{id} [get]
func (h *handler) GetFosterChildrenTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	res := h.service.GetFosterChildrenTransactionByID(ctx, id)
	c.JSON(res.Status, res)
}

// CreateOfflineFosterChildrenTransaction
//
// @Summary Create Offline Foster Children Transaction
// @Description Create a foster children transaction without initiating a Midtrans payment (admin only)
// @Tags Foster Children
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Foster Children ID"
// @Param body body CreateFosterChildrenTransactionRequest true "Offline transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/admin/foster-children/{id}/transactions [post]
func (h *handler) CreateOfflineFosterChildrenTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	fosterChildrenID := c.Param("id")

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var req CreateFosterChildrenTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}
	res := h.service.CreateOfflineFosterChildrenTransaction(ctx, claims.AccountID, fosterChildrenID, req)
	c.JSON(res.Status, res)
}

// CreateFosterChildrenTransaction
//
// @Summary Create Foster Children Transaction
// @Description Initiate a Midtrans Snap payment for a foster children donation
// @Tags Foster Children
// @Accept json
// @Produce json
// @Param id path string true "Foster Children ID"
// @Param body body CreateFosterChildrenTransactionRequest true "Transaction request"
// @Success 201 {object} pkg.Response
// @Router /api/foster-children/{id}/transactions [post]
func (h *handler) CreateFosterChildrenTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	fosterChildrenID := c.Param("id")
	accountID := ""
	if userData, exists := c.Get("user_data"); exists {
		if claims, ok := userData.(jwt_pkg.UserJWTClaims); ok {
			accountID = claims.AccountID
		}
	}
	var req CreateFosterChildrenTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid request body", nil, nil))
		return
	}

	res := h.service.CreateFosterChildrenTransaction(ctx, accountID, fosterChildrenID, req)
	c.JSON(res.Status, res)
}

// GetMyFosterChildrenTransactionList
//
// @Summary List My Foster Children Transactions
// @Description Retrieve a paginated list of foster children transactions for the authenticated user
// @Tags Foster Children
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by payment status"
// @Param limit query int false "Items per page"
// @Param next_cursor query string false "Cursor for next page"
// @Param prev_cursor query string false "Cursor for previous page"
// @Success 200 {object} pkg.Response{data=FosterChildrenTransactionListResponse}
// @Router /api/me/foster-children/transactions [get]
func (h *handler) GetMyFosterChildrenTransactionList(c *gin.Context) {
	ctx := c.Request.Context()

	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	var params FosterChildrenTransactionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(http.StatusBadRequest, "Invalid query parameters", nil, nil))
		return
	}

	res := h.service.GetMyFosterChildrenTransactionList(ctx, claims.AccountID, params)
	c.JSON(res.Status, res)
}

// GetMyFosterChildrenTransactionByID
//
// @Summary Get My Foster Children Transaction by ID
// @Description Retrieve a specific foster children transaction owned by the authenticated user
// @Tags Foster Children
// @Security BearerAuth
// @Produce json
// @Param id path string true "Transaction ID"
// @Success 200 {object} pkg.Response
// @Router /api/me/foster-children/transactions/{id} [get]
func (h *handler) GetMyFosterChildrenTransactionByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	claims := c.MustGet("user_data").(jwt_pkg.UserJWTClaims)

	res := h.service.GetMyFosterChildrenTransactionByID(ctx, id, claims.AccountID)
	c.JSON(res.Status, res)
}
