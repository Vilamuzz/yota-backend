package donation_program

import (
	"mime/multipart"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type DonationProgramRequest struct {
	Title       string                `json:"title" form:"title"`
	CoverImage  *multipart.FileHeader `json:"coverImage" form:"coverImage" swaggerignore:"true"`
	Category    Category              `json:"category" form:"category"`
	Description string                `json:"description" form:"description"`
	FundTarget  float64               `json:"fundTarget" form:"fundTarget"`
	Status      Status                `json:"status" form:"status"`
	StartDate   time.Time             `json:"startDate" form:"startDate" time_format:"2006-01-02"`
	EndDate     time.Time             `json:"endDate" form:"endDate" time_format:"2006-01-02"`
}
type DonationProgramQueryParams struct {
	Search   string   `form:"search"`
	Category Category `form:"category"`
	Status   string   `form:"status"`
	pkg.PaginationParams
}
