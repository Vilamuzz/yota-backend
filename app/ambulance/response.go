package ambulance

import "github.com/Vilamuzz/yota-backend/pkg"

type AmbulanceResponse struct {
	ID          string `json:"id"`
	PlateNumber string `json:"plate_number"`
	Phone       string `json:"phone"`
}

type AmbulanceListResponse struct {
	Ambulances []AmbulanceResponse  `json:"ambulances"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (a *Ambulance) toAmbulanceResponse() AmbulanceResponse {
	return AmbulanceResponse{
		ID:          a.ID.String(),
		PlateNumber: a.PlateNumber,
		Phone:       a.Phone,
	}
}

func toAmbulanceListResponse(ambulances []Ambulance, pagination pkg.CursorPagination) AmbulanceListResponse {
	var responses []AmbulanceResponse
	for _, ambulance := range ambulances {
		responses = append(responses, ambulance.toAmbulanceResponse())
	}
	if responses == nil {
		responses = []AmbulanceResponse{}
	}
	return AmbulanceListResponse{
		Ambulances: responses,
		Pagination: pagination,
	}
}
