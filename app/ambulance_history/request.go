package ambulance_history

import "github.com/Vilamuzz/yota-backend/pkg"

type CreateAmbulanceHistoryRequest struct {
	AmbulanceID     string          `json:"ambulanceId"`
	DriverID        string          `json:"driverId"`
	ServiceCategory ServiceCategory `json:"serviceCategory"`
	Note            string          `json:"note"`
}

type UpdateAmbulanceHistoryRequest struct {
	ServiceCategory ServiceCategory `json:"serviceCategory"`
}

type AmbulanceHistoryQueryParams struct {
	AmbulanceID     string `form:"ambulanceId"`
	ServiceCategory string `form:"serviceCategory"`
	pkg.PaginationParams
}

type AmbulanceSummaryQueryParams struct {
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}
