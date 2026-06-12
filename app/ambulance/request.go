package ambulance

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateAmbulanceRequest struct {
	DriverID    string                `form:"driverId"`
	Image       *multipart.FileHeader `form:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber"`
	Status      AmbulanceStatus       `form:"status"`
}

type UpdateAmbulanceRequest struct {
	DriverID    string                `form:"driverId"`
	Image       *multipart.FileHeader `form:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber"`
	Status      AmbulanceStatus       `form:"status"`
}

type AmbulanceQueryParams struct {
	Search      string          `json:"search" form:"search"`
	Status      AmbulanceStatus `json:"status" form:"status"`
	DriverID    string          `json:"-" form:"-"`
	AmbulanceID string          `json:"-" form:"-"`
	pkg.CursorPagination
}
