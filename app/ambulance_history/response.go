package ambulance_history

import (
	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type CategoryCount struct {
	Category ServiceCategory `json:"category"`
	Count    int64           `json:"count"`
}

type SummaryResponse struct {
	Total      int64           `json:"total"`
	Categories []CategoryCount `json:"categories"`
	StartDate  string          `json:"startDate"`
	EndDate    string          `json:"endDate"`
}

type HistoryResponse struct {
	Driver          account.DriverResponse `json:"driver"`
	ServiceCategory ServiceCategory        `json:"serviceCategory"`
	CreatedAt       string                 `json:"createdAt"`
}

type HistoryListResponse struct {
	Histories  []HistoryResponse    `json:"histories"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

type AdminHistoryResponse struct {
	ID              string                 `json:"id"`
	Driver          account.DriverResponse `json:"driver"`
	ServiceCategory ServiceCategory        `json:"serviceCategory"`
	Note            string                 `json:"note"`
	CreatedAt       string                 `json:"createdAt"`
}

type AdminHistoryListResponse struct {
	Histories  []AdminHistoryResponse `json:"histories"`
	Pagination pkg.CursorPagination   `json:"pagination"`
}

func (h *AmbulanceHistory) toAmbulanceHistoryToResponse() HistoryResponse {
	driver := account.DriverResponse{
		ID:       h.DriverID.String(),
		Username: "Unknown",
		Phone:    "-",
	}
	if h.Driver.UserProfile.ID != uuid.Nil {
		driver.Username = h.Driver.UserProfile.Username
		if h.Driver.UserProfile.Phone != nil {
			driver.Phone = *h.Driver.UserProfile.Phone
		}
	}

	return HistoryResponse{
		Driver:          driver,
		ServiceCategory: h.ServiceCategory,
		CreatedAt:       h.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAmbulanceHistoriesToListResponse(histories []AmbulanceHistory, pagination pkg.CursorPagination) HistoryListResponse {
	var responses []HistoryResponse
	for _, history := range histories {
		responses = append(responses, history.toAmbulanceHistoryToResponse())
	}
	if histories == nil {
		responses = []HistoryResponse{}
	}
	return HistoryListResponse{
		Histories:  responses,
		Pagination: pagination,
	}
}

func (h *AmbulanceHistory) toAdminAmbulanceHistoryToResponse() AdminHistoryResponse {
	driver := account.DriverResponse{
		ID:       h.DriverID.String(),
		Username: "Unknown",
		Phone:    "-",
	}
	if h.Driver.UserProfile.ID != uuid.Nil {
		driver.Username = h.Driver.UserProfile.Username
		if h.Driver.UserProfile.Phone != nil {
			driver.Phone = *h.Driver.UserProfile.Phone
		}
	}

	return AdminHistoryResponse{
		ID:              h.ID.String(),
		Driver:          driver,
		ServiceCategory: h.ServiceCategory,
		Note:            h.Note,
		CreatedAt:       h.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toAmbulanceHistoriesToAdminListResponse(histories []AmbulanceHistory, pagination pkg.CursorPagination) AdminHistoryListResponse {
	var responses []AdminHistoryResponse
	for _, history := range histories {
		responses = append(responses, history.toAdminAmbulanceHistoryToResponse())
	}
	if histories == nil {
		responses = []AdminHistoryResponse{}
	}
	return AdminHistoryListResponse{
		Histories:  responses,
		Pagination: pagination,
	}
}

type HistoryMonthlyTrendItem struct {
	Month            string `json:"month"`
	SocialService    int    `json:"socialService"`
	MortuaryService  int    `json:"mortuaryService"`
	PatientService   int    `json:"patientService"`
	EmergencyService int    `json:"emergencyService"`
	OtherService     int    `json:"otherService"`
}

type HistoryMonthlyTrendRecord struct {
	Year  string                    `json:"year"`
	Items []HistoryMonthlyTrendItem `json:"items"`
}
