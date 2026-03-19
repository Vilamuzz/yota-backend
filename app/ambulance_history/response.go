package ambulance_history

import "github.com/Vilamuzz/yota-backend/pkg"

type HistoryResponse struct {
	ID              int             `json:"id"`
	AmbulanceID     int             `json:"ambulance_id"`
	UserID          int             `json:"user_id"`
	ServiceCategory ServiceCategory `json:"service_category"`
	CreatedAt       string          `json:"created_at"`
}

type HistoryListResponse struct {
	Histories  []HistoryResponse    `json:"histories"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (h *AmbulanceHistory) toAmbulanceHistoryToResponse() HistoryResponse {
	return HistoryResponse{
		ID:              h.ID,
		AmbulanceID:     h.AmbulanceID,
		UserID:          h.UserID,
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
