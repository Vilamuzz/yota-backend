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

// SummaryPeriod defines the time period for the summary.
// Accepted values: all_time, this_week, this_month, this_year, custom
// When using "custom", StartDate and EndDate must be provided in YYYY-MM-DD format.
type SummaryPeriod string

const (
	PeriodAllTime   SummaryPeriod = "all_time"
	PeriodThisWeek  SummaryPeriod = "this_week"
	PeriodThisMonth SummaryPeriod = "this_month"
	PeriodThisYear  SummaryPeriod = "this_year"
	PeriodCustom    SummaryPeriod = "custom"
)

type AmbulanceSummaryQueryParams struct {
	// Period is the time window for the summary.
	// One of: all_time, this_week, this_month, this_year, custom (default: all_time)
	Period SummaryPeriod `form:"period"`
	// Date is used only when Period = "custom" (format: YYYY-MM-DD)
	StartDate string `form:"startDate"`
	EndDate   string `form:"end_date"`
}
