package donation_program

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	s3_pkg "github.com/Vilamuzz/yota-backend/pkg/s3"
)

type DonationProgramResponse struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImage    string    `json:"coverImage"`
	Category      Category  `json:"category"`
	Description   string    `json:"description"`
	FundTarget    float64   `json:"fundTarget"`
	CollectedFund float64   `json:"collectedFund"`
	TotalExpense  float64   `json:"totalExpense"`
	Status        Status    `json:"status"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
}

type AdminDonationProgramResponse struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CoverImage    string    `json:"coverImage"`
	Category      Category  `json:"category"`
	FundTarget    float64   `json:"fundTarget"`
	CollectedFund float64   `json:"collectedFund"`
	TotalExpense  float64   `json:"totalExpense"`
	Status        Status    `json:"status"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AdminDonationProgramListResponse struct {
	DonationPrograms []AdminDonationProgramResponse `json:"donationPrograms"`
	Pagination       pkg.OffsetPagination           `json:"pagination"`
}

type DonationProgramListResponse struct {
	DonationPrograms []DonationProgramResponse `json:"donationPrograms"`
	Pagination       pkg.OffsetPagination      `json:"pagination"`
}

func (d *DonationProgram) toAdminDonationProgramResponse() AdminDonationProgramResponse {
	return AdminDonationProgramResponse{
		ID:            d.ID.String(),
		Title:         d.Title,
		Description:   d.Description,
		CoverImage:    s3_pkg.GetCDNURL(d.CoverImage),
		Category:      d.Category,
		FundTarget:    d.FundTarget,
		CollectedFund: d.CollectedFund,
		TotalExpense:  d.TotalExpense,
		Status:        d.Status,
		StartDate:     d.StartDate,
		EndDate:       d.EndDate,
		CreatedAt:     d.CreatedAt,
	}
}

func (d *DonationProgram) toDonationProgramResponse() DonationProgramResponse {
	return DonationProgramResponse{
		ID:            d.ID.String(),
		Title:         d.Title,
		Slug:          d.Slug,
		Description:   d.Description,
		CoverImage:    s3_pkg.GetCDNURL(d.CoverImage),
		Category:      d.Category,
		FundTarget:    d.FundTarget,
		CollectedFund: d.CollectedFund,
		TotalExpense:  d.TotalExpense,
		Status:        d.Status,
		StartDate:     d.StartDate,
		EndDate:       d.EndDate,
	}
}

func toAdminDonationProgramListResponse(donations []DonationProgram, pagination pkg.OffsetPagination) AdminDonationProgramListResponse {
	var responses []AdminDonationProgramResponse
	for _, d := range donations {
		responses = append(responses, d.toAdminDonationProgramResponse())
	}

	if responses == nil {
		responses = []AdminDonationProgramResponse{}
	}

	return AdminDonationProgramListResponse{
		DonationPrograms: responses,
		Pagination:       pagination,
	}
}

func toDonationProgramListResponse(donations []DonationProgram, pagination pkg.OffsetPagination) DonationProgramListResponse {
	var responses []DonationProgramResponse
	for _, d := range donations {
		responses = append(responses, d.toDonationProgramResponse())
	}

	if responses == nil {
		responses = []DonationProgramResponse{}
	}

	return DonationProgramListResponse{
		DonationPrograms: responses,
		Pagination:       pagination,
	}
}
