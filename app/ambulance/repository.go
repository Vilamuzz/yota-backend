package ambulance

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	// Ambulance CRUD
	CreateAmbulance(ctx context.Context, ambulance *Ambulance) error
	FetchAmbulance(ctx context.Context, id uuid.UUID) (*Ambulance, error)
	FetchAmbulanceByPlate(ctx context.Context, plateNumber string) (*Ambulance, error)
	FetchAllAmbulances(ctx context.Context) ([]Ambulance, error)
	UpdateAmbulance(ctx context.Context, ambulance *Ambulance) error
	UpdateAmbulanceLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error
	UpdateAmbulanceStatus(ctx context.Context, id uuid.UUID, status AmbulanceStatus) error

	// Tracking Session
	CreateTrackingSession(ctx context.Context, session *TrackingSession) error
	EndTrackingSession(ctx context.Context, sessionID uuid.UUID) error
	GetActiveSession(ctx context.Context, ambulanceID uuid.UUID) (*TrackingSession, error)

	// Location History
	SaveLocationHistory(ctx context.Context, history *LocationHistory) error
	GetLocationHistory(ctx context.Context, ambulanceID uuid.UUID, sessionID uuid.UUID) ([]LocationHistory, error)
	GetRecentLocations(ctx context.Context, ambulanceID uuid.UUID, limit int) ([]LocationHistory, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) CreateAmbulance(ctx context.Context, ambulance *Ambulance) error {
	return r.Conn.WithContext(ctx).Create(ambulance).Error
}

func (r *repository) FetchAmbulance(ctx context.Context, id uuid.UUID) (*Ambulance, error) {
	var ambulance Ambulance
	if err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&ambulance).Error; err != nil {
		return nil, err
	}
	return &ambulance, nil
}

func (r *repository) FetchAmbulanceByPlate(ctx context.Context, plateNumber string) (*Ambulance, error) {
	var ambulance Ambulance
	if err := r.Conn.WithContext(ctx).Where("plate_number = ?", plateNumber).First(&ambulance).Error; err != nil {
		return nil, err
	}
	return &ambulance, nil
}

func (r *repository) FetchAllAmbulances(ctx context.Context) ([]Ambulance, error) {
	var ambulances []Ambulance
	if err := r.Conn.WithContext(ctx).Find(&ambulances).Error; err != nil {
		return nil, err
	}
	return ambulances, nil
}

func (r *repository) UpdateAmbulance(ctx context.Context, ambulance *Ambulance) error {
	return r.Conn.WithContext(ctx).Save(ambulance).Error
}

func (r *repository) UpdateAmbulanceLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error {
	return r.Conn.WithContext(ctx).Model(&Ambulance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"current_lat":    lat,
			"current_lng":    lng,
			"last_update_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *repository) UpdateAmbulanceStatus(ctx context.Context, id uuid.UUID, status AmbulanceStatus) error {
	return r.Conn.WithContext(ctx).Model(&Ambulance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *repository) CreateTrackingSession(ctx context.Context, session *TrackingSession) error {
	return r.Conn.WithContext(ctx).Create(session).Error
}

func (r *repository) EndTrackingSession(ctx context.Context, sessionID uuid.UUID) error {
	return r.Conn.WithContext(ctx).Model(&TrackingSession{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"ended_at": gorm.Expr("NOW()"),
			"status":   "ended",
		}).Error
}

func (r *repository) GetActiveSession(ctx context.Context, ambulanceID uuid.UUID) (*TrackingSession, error) {
	var session TrackingSession
	if err := r.Conn.WithContext(ctx).
		Where("ambulance_id = ? AND status = ?", ambulanceID, "active").
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repository) SaveLocationHistory(ctx context.Context, history *LocationHistory) error {
	return r.Conn.WithContext(ctx).Create(history).Error
}

func (r *repository) GetLocationHistory(ctx context.Context, ambulanceID uuid.UUID, sessionID uuid.UUID) ([]LocationHistory, error) {
	var history []LocationHistory
	if err := r.Conn.WithContext(ctx).
		Where("ambulance_id = ? AND session_id = ?", ambulanceID, sessionID).
		Order("recorded_at ASC").
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}

func (r *repository) GetRecentLocations(ctx context.Context, ambulanceID uuid.UUID, limit int) ([]LocationHistory, error) {
	var history []LocationHistory
	if err := r.Conn.WithContext(ctx).
		Where("ambulance_id = ?", ambulanceID).
		Order("recorded_at DESC").
		Limit(limit).
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}
