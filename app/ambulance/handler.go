package ambulance

import (
	"net/http"

	"github.com/Vilamuzz/yota-backend/app/middleware"
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

type handler struct {
	service    Service
	hub        *Hub
	middleware middleware.AppMiddleware
}

func NewHandler(r *gin.RouterGroup, s Service, h *Hub, m middleware.AppMiddleware) {
	handler := &handler{
		service:    s,
		hub:        h,
		middleware: m,
	}
	handler.RegisterRoutes(r)
}

func (h *handler) RegisterRoutes(r *gin.RouterGroup) {
	api := r.Group("/ambulance")

	// WebSocket endpoints
	api.GET("/ws/mobile", h.HandleMobileWebSocket)
	api.GET("/ws/web", h.HandleWebWebSocket)

	// REST endpoints
	protected := api.Group("")
	protected.Use(h.middleware.RequireRoles(
		string(user.RoleAmbulanceManager),
		string(user.RoleSuperadmin),
	))
	{
		protected.POST("/", h.CreateAmbulance)
		protected.GET("/", h.GetAllAmbulances)
		protected.GET("/:id", h.GetAmbulance)
		protected.GET("/:id/history/:session_id", h.GetTrackingHistory)
		protected.POST("/:id/tracking/start", h.StartTracking)
		protected.POST("/:id/tracking/stop", h.StopTracking)
	}

	// Public endpoints for online status
	api.GET("/online", h.GetOnlineAmbulances)
}

// HandleMobileWebSocket handles WebSocket connections from mobile app
// @Summary Mobile WebSocket Connection
// @Description WebSocket endpoint for ambulance driver mobile app
// @Tags Ambulance
// @Param ambulance_id query string true "Ambulance ID"
// @Param token query string true "JWT Token"
// @Router /api/ambulance/ws/mobile [get]
func (h *handler) HandleMobileWebSocket(c *gin.Context) {
	ambulanceID := c.Query("ambulance_id")
	token := c.Query("token")

	if ambulanceID == "" || token == "" {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(
			http.StatusBadRequest,
			"ambulance_id and token are required",
			nil, nil,
		))
		return
	}

	// TODO: Validate token here
	// claims, err := validateToken(token)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ID:          uuid.New().String(),
		Conn:        conn,
		Hub:         h.hub,
		Send:        make(chan []byte, 256),
		ClientType:  ClientTypeMobile,
		AmbulanceID: ambulanceID,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump(h.service)
}

// HandleWebWebSocket handles WebSocket connections from web dashboard
// @Summary Web WebSocket Connection
// @Description WebSocket endpoint for admin web dashboard
// @Tags Ambulance
// @Param token query string true "JWT Token"
// @Router /api/ambulance/ws/web [get]
func (h *handler) HandleWebWebSocket(c *gin.Context) {
	token := c.Query("token")

	if token == "" {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(
			http.StatusBadRequest,
			"token is required",
			nil, nil,
		))
		return
	}

	// TODO: Validate token and check permissions

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		ID:         uuid.New().String(),
		Conn:       conn,
		Hub:        h.hub,
		Send:       make(chan []byte, 256),
		ClientType: ClientTypeWeb,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump(h.service)
}

// CreateAmbulance creates a new ambulance
// @Summary Create Ambulance
// @Description Create a new ambulance record
// @Tags Ambulance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body CreateAmbulanceRequest true "Ambulance Data"
// @Success 201 {object} pkg.Response
// @Router /api/ambulance [post]
func (h *handler) CreateAmbulance(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateAmbulanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, pkg.NewResponse(
			http.StatusBadRequest,
			"Invalid request",
			nil, nil,
		))
		return
	}

	res := h.service.CreateAmbulance(ctx, req)
	c.JSON(res.Status, res)
}

// GetAllAmbulances retrieves all ambulances
// @Summary Get All Ambulances
// @Description Retrieve all ambulances with online status
// @Tags Ambulance
// @Security BearerAuth
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/ambulance [get]
func (h *handler) GetAllAmbulances(c *gin.Context) {
	ctx := c.Request.Context()
	res := h.service.GetAllAmbulances(ctx)
	c.JSON(res.Status, res)
}

// GetAmbulance retrieves a specific ambulance
// @Summary Get Ambulance
// @Description Retrieve ambulance by ID
// @Tags Ambulance
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ambulance ID"
// @Success 200 {object} pkg.Response
// @Router /api/ambulance/{id} [get]
func (h *handler) GetAmbulance(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.GetAmbulance(ctx, id)
	c.JSON(res.Status, res)
}

// StartTracking starts a tracking session
// @Summary Start Tracking
// @Description Start tracking session for an ambulance
// @Tags Ambulance
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ambulance ID"
// @Success 201 {object} pkg.Response
// @Router /api/ambulance/{id}/tracking/start [post]
func (h *handler) StartTracking(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.StartTracking(ctx, id)
	c.JSON(res.Status, res)
}

// StopTracking stops a tracking session
// @Summary Stop Tracking
// @Description Stop tracking session for an ambulance
// @Tags Ambulance
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ambulance ID"
// @Success 200 {object} pkg.Response
// @Router /api/ambulance/{id}/tracking/stop [post]
func (h *handler) StopTracking(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	res := h.service.StopTracking(ctx, id)
	c.JSON(res.Status, res)
}

// GetTrackingHistory retrieves tracking history
// @Summary Get Tracking History
// @Description Get location history for a tracking session
// @Tags Ambulance
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ambulance ID"
// @Param session_id path string true "Session ID"
// @Success 200 {object} pkg.Response
// @Router /api/ambulance/{id}/history/{session_id} [get]
func (h *handler) GetTrackingHistory(c *gin.Context) {
	ctx := c.Request.Context()
	ambulanceID := c.Param("id")
	sessionID := c.Param("session_id")
	res := h.service.GetTrackingHistory(ctx, ambulanceID, sessionID)
	c.JSON(res.Status, res)
}

// GetOnlineAmbulances retrieves online ambulances
// @Summary Get Online Ambulances
// @Description Get list of currently online ambulances
// @Tags Ambulance
// @Produce json
// @Success 200 {object} pkg.Response
// @Router /api/ambulance/online [get]
func (h *handler) GetOnlineAmbulances(c *gin.Context) {
	onlineIDs := h.service.GetOnlineAmbulances()
	c.JSON(http.StatusOK, pkg.NewResponse(
		http.StatusOK,
		"Online ambulances retrieved",
		nil,
		map[string]interface{}{
			"count":         len(onlineIDs),
			"ambulance_ids": onlineIDs,
		},
	))
}
