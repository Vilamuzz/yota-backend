package prayer

import (
	"github.com/Vilamuzz/yota-backend/pkg"
)

type PrayerResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type PrayerListResponse struct {
	Prayers    []PrayerResponse     `json:"prayers"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (p *Prayer) toPrayerResponse() PrayerResponse {
	username := p.DonationProgramTransaction.DonorName
	if username == "" {
		username = "Anonymous"
	}

	return PrayerResponse{
		ID:        p.ID.String(),
		Username:  username,
		Content:   p.Content,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toPrayerListResponse(prayers []Prayer, pagination pkg.CursorPagination) PrayerListResponse {
	var responses []PrayerResponse
	for _, prayer := range prayers {
		responses = append(responses, prayer.toPrayerResponse())
	}
	if responses == nil {
		responses = []PrayerResponse{}
	}
	return PrayerListResponse{
		Prayers:    responses,
		Pagination: pagination,
	}
}

type PrayerReportedResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type PrayerReportedListResponse struct {
	Prayers    []PrayerReportedResponse `json:"prayers"`
	Pagination pkg.CursorPagination     `json:"pagination"`
}

func (p *Prayer) toPrayerReportedResponse() PrayerReportedResponse {
	username := p.DonationProgramTransaction.DonorName
	if username == "" {
		username = "Anonymous"
	}

	return PrayerReportedResponse{
		ID:        p.ID.String(),
		Username:  username,
		Content:   p.Content,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toPrayerReportedListResponse(prayers []Prayer, pagination pkg.CursorPagination) PrayerReportedListResponse {
	var responses []PrayerReportedResponse
	for _, prayer := range prayers {
		responses = append(responses, prayer.toPrayerReportedResponse())
	}
	if responses == nil {
		responses = []PrayerReportedResponse{}
	}
	return PrayerReportedListResponse{
		Prayers:    responses,
		Pagination: pagination,
	}
}
