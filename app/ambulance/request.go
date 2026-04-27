package ambulance

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateAmbulanceRequest struct {
	Image       *multipart.FileHeader `form:"image" json:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber" json:"plateNumber"`
	Phone       string                `form:"phone" json:"phone"`
}

type UpdateAmbulanceRequest struct {
	Image       *multipart.FileHeader `form:"image" json:"image" swaggerignore:"true"`
	PlateNumber string                `form:"plateNumber" json:"plateNumber"`
	Phone       string                `form:"phone" json:"phone"`
}

type AmbulanceQueryParams struct {
	Search string `json:"search"`
	pkg.CursorPagination
}
