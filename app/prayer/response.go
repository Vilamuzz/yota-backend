package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerResponse struct {
	ID         string `json:"id"`
	DonationID string `json:"donation_id"`
	UserID     string `json:"user_id"`
	Content    string `json:"content"`
	LikeCount  int    `json:"like_count"`
	CreatedAt  string `json:"created_at"`
}

type PrayerListResponse struct {
	Prayers    []PrayerResponse     `json:"prayers"`
	Pagination pkg.CursorPagination `json:"pagination"`
}

func (p *Prayer) toPrayerResponse() PrayerResponse {
	return PrayerResponse{
		ID:         p.ID,
		DonationID: p.DonationID,
		UserID:     p.UserID,
		Content:    p.Content,
		LikeCount:  p.LikeCount,
		CreatedAt:  p.CreatedAt.Format("2006-01-02 15:04:05"),
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
	DonationID  string `json:"donation_id"`
	UserID      string `json:"user_id"`
	Content     string `json:"content"`
	LikeCount   int    `json:"like_count"`
	ReportCount int    `json:"report_count"`
	CreatedAt   string `json:"created_at"`
}

type PrayerReportedListResponse struct {
	Prayers    []PrayerReportedResponse `json:"prayers"`
	Pagination pkg.CursorPagination     `json:"pagination"`
}

func (p *Prayer) toPrayerReportedResponse() PrayerReportedResponse {
	return PrayerReportedResponse{
		ID:          p.ID,
		DonationID:  p.DonationID,
		UserID:      p.UserID,
		Content:     p.Content,
		LikeCount:   p.LikeCount,
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
