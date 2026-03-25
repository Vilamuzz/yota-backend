package prayer

import (
	"github.com/Vilamuzz/yota-backend/app/user"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type PrayerResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	AmenCount int    `json:"amen_count"`
	IsAmen    bool   `json:"is_amen"`
	CreatedAt string `json:"created_at"`
}

type PrayerListResponse struct {
	Prayers    []PrayerResponse     `json:"prayers"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (p *Prayer) toPrayerResponse() PrayerResponse {
	if p.User == nil {
		p.User = &user.User{Username: "Anonymous"}
	}
	return PrayerResponse{
		ID:        p.ID,
		Username:  p.User.Username,
		Content:   p.Content,
		AmenCount: p.AmenCount,
		IsAmen:    p.IsAmen,
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
	ID          string `json:"id"`
	Username    string `json:"username"`
	Content     string `json:"content"`
	AmenCount   int    `json:"amen_count"`
	ReportCount int    `json:"report_count"`
	CreatedAt   string `json:"created_at"`
}

type PrayerReportedListResponse struct {
	Prayers    []PrayerReportedResponse `json:"prayers"`
	Pagination pkg.CursorPagination     `json:"pagination"`
}

func (p *Prayer) toPrayerReportedResponse() PrayerReportedResponse {
	if p.User == nil {
		p.User = &user.User{Username: "Anonymous"}
	}
	return PrayerReportedResponse{
		ID:          p.ID,
		Username:    p.User.Username,
		Content:     p.Content,
		AmenCount:   p.AmenCount,
		ReportCount: p.ReportCount,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
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
