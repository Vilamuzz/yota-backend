package prayer

import (
	"github.com/Vilamuzz/yota-backend/pkg"
)

type PrayerResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	IsAmen    bool   `json:"isAmen"`
	AmenCount int64  `json:"amenCount"`
	CreatedAt string `json:"createdAt"`
}

type AdminPrayerResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Content     string `json:"content"`
	CreatedAt   string `json:"createdAt"`
	AmenCount   int64  `json:"amenCount"`
	ReportCount int64  `json:"reportCount"`
}

type AdminPrayerListResponse struct {
	Prayers    []AdminPrayerResponse `json:"prayers"`
	Pagination pkg.OffsetPagination  `json:"pagination"`
}

type PrayerListResponse struct {
	Prayers    []PrayerResponse     `json:"prayers"`
	Pagination pkg.OffsetPagination `json:"pagination"`
}

func (p *Prayer) toPrayerResponse() PrayerResponse {
	username := p.DonationProgramTransaction.DonorName
	if username == "" {
		username = "Hamba Allah"
	}

	return PrayerResponse{
		ID:        p.ID.String(),
		Username:  username,
		Content:   p.Content,
		IsAmen:    p.IsAmen,
		AmenCount: p.AmenCount,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toPrayerListResponse(prayers []Prayer, pagination pkg.OffsetPagination) PrayerListResponse {
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

func (p *Prayer) toAdminPrayerResponse() AdminPrayerResponse {
	username := p.DonationProgramTransaction.DonorName
	if username == "" {
		username = "Hamba Allah"
	}

	return AdminPrayerResponse{
		ID:          p.ID.String(),
		Username:    username,
		Content:     p.Content,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		AmenCount:   p.AmenCount,
		ReportCount: p.ReportCount,
	}
}

func toAdminPrayerListResponse(prayers []Prayer, pagination pkg.OffsetPagination) AdminPrayerListResponse {
	var responses []AdminPrayerResponse
	for _, prayer := range prayers {
		responses = append(responses, prayer.toAdminPrayerResponse())
	}
	if responses == nil {
		responses = []AdminPrayerResponse{}
	}
	return AdminPrayerListResponse{
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
	Pagination pkg.OffsetPagination     `json:"pagination"`
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

func toPrayerReportedListResponse(prayers []Prayer, pagination pkg.OffsetPagination) PrayerReportedListResponse {
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
