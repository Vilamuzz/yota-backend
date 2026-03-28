package foster_child

import (
	"mime/multipart"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type CreateFosterChildRequest struct {
	Name      string               `json:"name"`
	Age       int                  `json:"age"`
	Gender    string               `json:"gender"`
	Status    bool                 `json:"status"` // true for not graduated, false for graduated
	Category  Category             `json:"category"`
	BirthDate string               `json:"birth_date"`
	Address   string               `json:"address"`
	ImageURL  multipart.FileHeader `json:"image_url"`
}

type UpdateFosterChildRequest struct {
	Name      string                `json:"name"`
	Age       int                   `json:"age"`
	Gender    string                `json:"gender"`
	Status    bool                  `json:"status"` // true for not graduated, false for graduated
	Category  Category              `json:"category"`
	BirthDate string                `json:"birth_date"`
	Address   string                `json:"address"`
	ImageURL  *multipart.FileHeader `json:"image_url,omitempty"`
}

type FosterChildQueryParams struct {
	Status   bool   `form:"status"`
	Category string `form:"category"`
	pkg.PaginationParams
}
