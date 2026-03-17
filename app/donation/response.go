package donation

import (
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
)

type PublishedDonationResponse struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Description   string     `json:"description"`
	ImageURL      string     `json:"image_url"`
	Category      Category   `json:"category"`
	FundTarget    float64    `json:"fund_target"`
	CollectedFund float64    `json:"collected_fund"`
	Status        Status     `json:"status"`
	DateEnd       time.Time  `json:"date_end"`
	PublishedAt   *time.Time `json:"published_at"`
}

type DonationResponse struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"image_url"`
	Category      Category  `json:"category"`
	FundTarget    float64   `json:"fund_target"`
	CollectedFund float64   `json:"collected_fund"`
	Status        Status    `json:"status"`
	DateEnd       time.Time `json:"date_end"`
	CreatedAt     time.Time `json:"created_at"`
}

type DonationListResponse struct {
	Donations  []DonationResponse   `json:"donations"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

type PublishedDonationListResponse struct {
	Donations  []PublishedDonationResponse `json:"donations"`
	Pagination pkg.CursorPagination        `json:"pagination"`
}

func (d *Donation) toDonationResponse() DonationResponse {
	return DonationResponse{
		ID:            d.ID,
		Title:         d.Title,
		Description:   d.Description,
		ImageURL:      d.ImageURL,
		Category:      d.Category,
		FundTarget:    d.FundTarget,
		CollectedFund: d.CollectedFund,
		Status:        d.Status,
		DateEnd:       d.DateEnd,
		CreatedAt:     d.CreatedAt,
	}
}

func (d *Donation) toPublishedDonationResponse() PublishedDonationResponse {
	return PublishedDonationResponse{
		ID:            d.ID,
		Title:         d.Title,
		Slug:          d.Slug,
		Description:   d.Description,
		ImageURL:      d.ImageURL,
		Category:      d.Category,
		FundTarget:    d.FundTarget,
		CollectedFund: d.CollectedFund,
		Status:        d.Status,
		DateEnd:       d.DateEnd,
		PublishedAt:   d.PublishedAt,
	}
}

func toDonationListResponse(donations []Donation, pagination pkg.CursorPagination) DonationListResponse {
	var donationResponses []DonationResponse
	for _, d := range donations {
		donationResponses = append(donationResponses, d.toDonationResponse())
	}

	if donationResponses == nil {
		donationResponses = []DonationResponse{}
	}

	return DonationListResponse{
		Donations:  donationResponses,
		Pagination: pagination,
	}
}

func toPublishedDonationListResponse(donations []Donation, pagination pkg.CursorPagination) PublishedDonationListResponse {
	var responses []PublishedDonationResponse
	for _, d := range donations {
		responses = append(responses, d.toPublishedDonationResponse())
	}
	if responses == nil {
		responses = []PublishedDonationResponse{}
	}
	return PublishedDonationListResponse{
		Donations:  responses,
		Pagination: pagination,
	}
}
