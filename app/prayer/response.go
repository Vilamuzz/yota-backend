package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerResponse struct {
	ID         string `json:"id"`
	DonationID string `json:"donation_id"`
	UserID     string `json:"user_id"`
	Content    string `json:"content"`
	LikeCount  int    `json:"like_count"`
	IsReported bool   `json:"is_reported"`
	CreatedAt  string `json:"created_at"`
}

type PrayerListResponse struct {
	Prayers    []PrayerResponse     `json:"prayers"`
	Pagination pkg.CursorPagination `json:"pagination"`
}
