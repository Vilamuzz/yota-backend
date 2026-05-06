package ambulance_history

import "github.com/Vilamuzz/yota-backend/pkg"

type HistoryResponse struct {
	ID              string          `json:"id"`
	AmbulanceID     string          `json:"ambulanceId"`
	DriverID        string          `json:"driverId"`
	ServiceCategory ServiceCategory `json:"serviceCategory"`
	CreatedAt       string          `json:"createdAt"`
}

type HistoryListResponse struct {
	Histories  []HistoryResponse    `json:"histories"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (h *AmbulanceHistory) toAmbulanceHistoryToResponse() HistoryResponse {
	return HistoryResponse{
		ID:              h.ID.String(),
		AmbulanceID:     h.AmbulanceID.String(),
		DriverID:        h.DriverID.String(),
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
