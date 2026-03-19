package ambulance_history

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceHistoryRequest struct {
	AmbulanceID     int             `json:"ambulance_id"`
	UserID          int             `json:"user_id"`
	ServiceCategory ServiceCategory `json:"service_category"`
}

type UpdateAmbulanceHistoryRequest struct {
	ServiceCategory ServiceCategory `json:"service_category"`
}

type AmbulanceHistoryQueryParams struct {
	AmbulanceID     int    `json:"ambulance_id"`
	ServiceCategory string `json:"service_category"`
	pkg.CursorPagination
}
