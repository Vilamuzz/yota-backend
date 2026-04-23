package donation_program

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramRequest struct {
	Title       string                `json:"title" form:"title"`
	CoverImage  *multipart.FileHeader `json:"cover_image" form:"cover_image" swaggerignore:"true"`
	Category    Category              `json:"category" form:"category"`
	Description string                `json:"description" form:"description"`
	FundTarget  float64               `json:"fund_target" form:"fund_target"`
	Status      Status                `json:"status" form:"status"`
	StartDate   time.Time             `json:"start_date" form:"start_date" time_format:"2006-01-02"`
	EndDate     time.Time             `json:"end_date" form:"end_date" time_format:"2006-01-02"`
}
type DonationProgramQueryParams struct {
	Search   string   `form:"search"`
	Category Category `form:"category"`
	Status   string   `form:"status"`
	pkg.PaginationParams
}
