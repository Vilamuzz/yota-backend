package ambulance

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateAmbulanceRequest struct {
	Image       *multipart.FileHeader `form:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber"`
	Phone       string                `form:"phone"`
}

type UpdateAmbulanceRequest struct {
	Image       *multipart.FileHeader `form:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber"`
	Phone       string                `form:"phone"`
}

type AmbulanceQueryParams struct {
	Search string `json:"search"`
	pkg.CursorPagination
}
