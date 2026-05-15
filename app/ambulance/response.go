package ambulance

import (
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type AmbulanceResponse struct {
	ID          string          `json:"id"`
	DriverName  string          `json:"driverName"`
	DriverPhone string          `json:"driverPhone"`
	Image       string          `json:"image"`
	PlateNumber string          `json:"plateNumber"`
	Status      AmbulanceStatus `json:"status"`
}

type AmbulanceListResponse struct {
	Ambulances []AmbulanceResponse  `json:"ambulances"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (a *Ambulance) toAmbulanceResponse() AmbulanceResponse {
	driverName := "Unknown"
	driverPhone := "-"

	if a.Driver.ID != uuid.Nil && a.Driver.UserProfile.ID != uuid.Nil {
		driverName = a.Driver.UserProfile.Username
		if a.Driver.UserProfile.Phone != nil {
			driverPhone = *a.Driver.UserProfile.Phone
		}
	}

	return AmbulanceResponse{
		ID:          a.ID.String(),
		PlateNumber: a.PlateNumber,
		Image:       a.Image,
		DriverName:  driverName,
		DriverPhone: driverPhone,
		Status:      a.Status,
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
