package ambulance

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateAmbulanceRequest struct {
	Image       *multipart.FileHeader `json:"image"`
	PlateNumber string                `json:"plate_number"`
	Phone       string                `json:"phone"`
}

type UpdateAmbulanceRequest struct {
	Image       *multipart.FileHeader `json:"image"`
	PlateNumber string                `json:"plate_number"`
	Phone       string                `json:"phone"`
}

type AmbulanceQueryParams struct {
	Search string `json:"search"`
	pkg.CursorPagination
}
