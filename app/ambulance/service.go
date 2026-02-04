package ambulance

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type Service interface {
	// Ambulance management
	CreateAmbulance(ctx context.Context, req CreateAmbulanceRequest) pkg.Response
	GetAllAmbulances(ctx context.Context) pkg.Response
	GetAmbulance(ctx context.Context, id string) pkg.Response
	UpdateAmbulanceStatus(ambulanceID string, status AmbulanceStatus)

	// Tracking
	StartTracking(ctx context.Context, ambulanceID string) pkg.Response
	StopTracking(ctx context.Context, ambulanceID string) pkg.Response
	SaveLocationHistory(update LocationUpdate)
	GetTrackingHistory(ctx context.Context, ambulanceID string, sessionID string) pkg.Response

	// Real-time
	GetOnlineAmbulances() []string
	GetCurrentLocation(ambulanceID string) *LocationUpdate
}

type service struct {
	repo    Repository
	hub     *Hub
	timeout time.Duration
}

func NewService(repo Repository, hub *Hub, timeout time.Duration) Service {
	return &service{
		repo:    repo,
		hub:     hub,
		timeout: timeout,
	}
}

func (s *service) CreateAmbulance(ctx context.Context, req CreateAmbulanceRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambulance := &Ambulance{
		ID:          uuid.New(),
		PlateNumber: req.PlateNumber,
		DriverName:  req.DriverName,
		DriverPhone: req.DriverPhone,
		Status:      StatusOffline,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateAmbulance(ctx, ambulance); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to create ambulance", nil, nil)
	}

	return pkg.NewResponse(http.StatusCreated, "Ambulance created successfully", nil, ambulance)
}

func (s *service) GetAllAmbulances(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambulances, err := s.repo.FetchAllAmbulances(ctx)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch ambulances", nil, nil)
	}

	// Enrich with online status and current location
	onlineIDs := s.hub.GetOnlineAmbulances()
	onlineMap := make(map[string]bool)
	for _, id := range onlineIDs {
		onlineMap[id] = true
	}

	response := make([]AmbulanceResponse, len(ambulances))
	for i, amb := range ambulances {
		response[i] = AmbulanceResponse{
			ID:          amb.ID.String(),
			PlateNumber: amb.PlateNumber,
			DriverName:  amb.DriverName,
			DriverPhone: amb.DriverPhone,
			Status:      string(amb.Status),
			IsOnline:    onlineMap[amb.ID.String()],
			CurrentLat:  amb.CurrentLat,
			CurrentLng:  amb.CurrentLng,
		}

		// Get real-time location if online
		if loc := s.hub.GetAmbulanceLocation(amb.ID.String()); loc != nil {
			response[i].CurrentLat = loc.Latitude
			response[i].CurrentLng = loc.Longitude
		}
	}

	return pkg.NewResponse(http.StatusOK, "Ambulances retrieved successfully", nil, response)
}

func (s *service) GetAmbulance(ctx context.Context, id string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambulanceID, err := uuid.Parse(id)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid ambulance ID", nil, nil)
	}

	ambulance, err := s.repo.FetchAmbulance(ctx, ambulanceID)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Ambulance not found", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Ambulance retrieved successfully", nil, ambulance)
}

func (s *service) UpdateAmbulanceStatus(ambulanceID string, status AmbulanceStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	id, err := uuid.Parse(ambulanceID)
	if err != nil {
		return
	}

	s.repo.UpdateAmbulanceStatus(ctx, id, status)
}

func (s *service) StartTracking(ctx context.Context, ambulanceID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	id, err := uuid.Parse(ambulanceID)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid ambulance ID", nil, nil)
	}

	// Check if already has active session
	existing, _ := s.repo.GetActiveSession(ctx, id)
	if existing != nil {
		return pkg.NewResponse(http.StatusConflict, "Tracking session already active", nil, existing)
	}

	session := &TrackingSession{
		ID:          uuid.New(),
		AmbulanceID: id,
		StartedAt:   time.Now(),
		Status:      "active",
	}

	if err := s.repo.CreateTrackingSession(ctx, session); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to start tracking", nil, nil)
	}

	// Update ambulance status
	s.repo.UpdateAmbulanceStatus(ctx, id, StatusOnDuty)

	return pkg.NewResponse(http.StatusCreated, "Tracking started", nil, session)
}

func (s *service) StopTracking(ctx context.Context, ambulanceID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	id, err := uuid.Parse(ambulanceID)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid ambulance ID", nil, nil)
	}

	session, err := s.repo.GetActiveSession(ctx, id)
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "No active tracking session", nil, nil)
	}

	if err := s.repo.EndTrackingSession(ctx, session.ID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to stop tracking", nil, nil)
	}

	// Update ambulance status
	s.repo.UpdateAmbulanceStatus(ctx, id, StatusAvailable)

	return pkg.NewResponse(http.StatusOK, "Tracking stopped", nil, nil)
}

func (s *service) SaveLocationHistory(update LocationUpdate) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	ambulanceID, err := uuid.Parse(update.AmbulanceID)
	if err != nil {
		return
	}

	// Get active session
	session, err := s.repo.GetActiveSession(ctx, ambulanceID)
	if err != nil {
		return
	}

	history := &LocationHistory{
		ID:          uuid.New(),
		AmbulanceID: ambulanceID,
		SessionID:   session.ID,
		Latitude:    update.Latitude,
		Longitude:   update.Longitude,
		Speed:       update.Speed,
		Heading:     update.Heading,
		RecordedAt:  time.Unix(update.Timestamp, 0),
	}

	s.repo.SaveLocationHistory(ctx, history)

	// Also update ambulance's current location
	s.repo.UpdateAmbulanceLocation(ctx, ambulanceID, update.Latitude, update.Longitude)
}

func (s *service) GetTrackingHistory(ctx context.Context, ambulanceID string, sessionID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	ambID, err := uuid.Parse(ambulanceID)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid ambulance ID", nil, nil)
	}

	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Invalid session ID", nil, nil)
	}

	history, err := s.repo.GetLocationHistory(ctx, ambID, sessID)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to fetch history", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "History retrieved", nil, history)
}

func (s *service) GetOnlineAmbulances() []string {
	return s.hub.GetOnlineAmbulances()
}

func (s *service) GetCurrentLocation(ambulanceID string) *LocationUpdate {
	return s.hub.GetAmbulanceLocation(ambulanceID)
}
