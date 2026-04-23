package ambulance_history

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceHistoryRequest struct {
	AmbulanceID     string          `json:"ambulance_id"`
	DriverID        string          `json:"driver_id"`
	ServiceCategory ServiceCategory `json:"service_category"`
	Note            string          `json:"note"`
}

type UpdateAmbulanceHistoryRequest struct {
	ServiceCategory ServiceCategory `json:"service_category"`
}

type AmbulanceHistoryQueryParams struct {
	AmbulanceID     int    `json:"ambulance_id"`
	ServiceCategory string `json:"service_category"`
	pkg.CursorPagination
}
