package donation_program

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type PublishedDonationProgramResponse struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	CoverImage    string     `json:"cover_image"`
	Category      Category   `json:"category"`
	Description   string     `json:"description"`
	FundTarget    float64    `json:"fund_target"`
	CollectedFund float64    `json:"collected_fund"`
	Status        Status     `json:"status"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       time.Time  `json:"end_date"`
	PublishedAt   *time.Time `json:"published_at"`
}

type DonationProgramResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	CoverImage  string     `json:"cover_image"`
	Category    Category   `json:"category"`
	FundTarget  float64    `json:"fund_target"`
	Status      Status     `json:"status"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DonationProgramListResponse struct {
	Donations  []DonationProgramResponse `json:"donations"`
	Pagination pkg.CursorPagination      `json:"pagination"`
}

type PublishedDonationProgramListResponse struct {
	Donations  []PublishedDonationProgramResponse `json:"donations"`
	Pagination pkg.CursorPagination               `json:"pagination"`
}

func (d *DonationProgram) toDonationProgramResponse() DonationProgramResponse {
	return DonationProgramResponse{
		ID:          d.ID.String(),
		Title:       d.Title,
		Slug:        d.Slug,
		Description: d.Description,
		CoverImage:  d.CoverImage,
		Category:    d.Category,
		FundTarget:  d.FundTarget,
		Status:      d.Status,
		StartDate:   d.StartDate,
		EndDate:     d.EndDate,
		PublishedAt: d.PublishedAt,
		CreatedAt:   d.CreatedAt,
	}
}

func (d *DonationProgram) toPublishedDonationProgramResponse() PublishedDonationProgramResponse {
	return PublishedDonationProgramResponse{
		ID:          d.ID.String(),
		Title:       d.Title,
		Slug:        d.Slug,
		Description: d.Description,
		CoverImage:  d.CoverImage,
		Category:      d.Category,
		FundTarget:    d.FundTarget,
		CollectedFund: d.CollectedFund,
		Status:        d.Status,
		StartDate:     d.StartDate,
		EndDate:     d.EndDate,
		PublishedAt: d.PublishedAt,
	}
}

func toDonationProgramListResponse(donations []DonationProgram, pagination pkg.CursorPagination) DonationProgramListResponse {
	var responses []DonationProgramResponse
	for _, d := range donations {
		responses = append(responses, d.toDonationProgramResponse())
	}

	if responses == nil {
		responses = []DonationProgramResponse{}
	}

	return DonationProgramListResponse{
		Donations:  responses,
		Pagination: pagination,
	}
}

func toPublishedDonationProgramListResponse(donations []DonationProgram, pagination pkg.CursorPagination) PublishedDonationProgramListResponse {
	var responses []PublishedDonationProgramResponse
	for _, d := range donations {
		responses = append(responses, d.toPublishedDonationProgramResponse())
	}
	if responses == nil {
		responses = []PublishedDonationProgramResponse{}
	}
	return PublishedDonationProgramListResponse{
		Donations:  responses,
		Pagination: pagination,
	}
}
