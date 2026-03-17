package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type CreatePrayerRequest struct {
	DonationID string `json:"donation_id"`
	UserID     string `json:"user_id"`
	Content    string `json:"content"`
	LikeCount  int    `json:"like_count"`
	Reported   bool   `json:"reported"`
}

type PrayerQueryParams struct {
	DonationID string `form:"donation_id"`
	Reported   bool   `form:"reported"`
	pkg.PaginationParams
}
