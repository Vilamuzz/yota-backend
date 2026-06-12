package ambulance

import (
	"github.com/Vilamuzz/yota-backend/app/account"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
)

type AmbulanceResponse struct {
	ID          string                 `json:"id"`
	Driver      account.DriverResponse `json:"driver"`
	Image       string                 `json:"image"`
	PlateNumber string                 `json:"plateNumber"`
	Status      AmbulanceStatus        `json:"status"`
}

type AmbulanceListResponse struct {
	Ambulances []AmbulanceResponse  `json:"ambulances"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (a *Ambulance) toAmbulanceResponse() AmbulanceResponse {
	driver := account.DriverResponse{
		ID:       "",
		Username: "Unknown",
		Phone:    "-",
	}

	if a.Driver.ID != uuid.Nil && a.Driver.UserProfile.ID != uuid.Nil {
		driver.ID = a.DriverID.String()
		driver.Username = a.Driver.UserProfile.Username
		if a.Driver.UserProfile.Phone != nil {
			driver.Phone = *a.Driver.UserProfile.Phone
		}
	} else if a.DriverID != uuid.Nil {
		driver.ID = a.DriverID.String()
	}

	return AmbulanceResponse{
		ID:          a.ID.String(),
		Driver:      driver,
		Image:       a.Image,
		PlateNumber: a.PlateNumber,
		Status:      a.Status,
	}
}

func (a *Ambulance) ToAmbulanceResponse() AmbulanceResponse {
	return a.toAmbulanceResponse()
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
