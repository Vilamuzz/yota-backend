package ambulance

type CreateAmbulanceRequest struct {
	PlateNumber string `json:"plate_number" binding:"required"`
	DriverName  string `json:"driver_name" binding:"required"`
	DriverPhone string `json:"driver_phone" binding:"required"`
}

type UpdateAmbulanceRequest struct {
	PlateNumber string          `json:"plate_number,omitempty"`
	DriverName  string          `json:"driver_name,omitempty"`
	DriverPhone string          `json:"driver_phone,omitempty"`
	Status      AmbulanceStatus `json:"status,omitempty"`
}

type ConnectRequest struct {
	AmbulanceID string `json:"ambulance_id" binding:"required"`
	Token       string `json:"token" binding:"required"`
}
