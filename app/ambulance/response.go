package ambulance

type AmbulanceResponse struct {
	ID          string  `json:"id"`
	PlateNumber string  `json:"plate_number"`
	DriverName  string  `json:"driver_name"`
	DriverPhone string  `json:"driver_phone"`
	Status      string  `json:"status"`
	IsOnline    bool    `json:"is_online"`
	CurrentLat  float64 `json:"current_lat,omitempty"`
	CurrentLng  float64 `json:"current_lng,omitempty"`
}

type TrackingResponse struct {
	SessionID   string           `json:"session_id"`
	AmbulanceID string           `json:"ambulance_id"`
	Status      string           `json:"status"`
	Locations   []LocationUpdate `json:"locations,omitempty"`
}
