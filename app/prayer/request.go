package prayer

type CreatePrayerRequest struct {
	DonationID string `json:"donation_id"`
	UserID     string `json:"user_id"`
	Content    string `json:"content"`
	LikeCount  int    `json:"like_count"`
	IsReported bool   `json:"is_reported"`
}

type PrayerQueryParams struct {
	DonationID string `form:"donation_id"`
	IsReported bool   `form:"is_reported"`
	NextCursor string `form:"next_cursor"`
	PrevCursor string `form:"prev_cursor"`
	Limit      int    `form:"limit"`
}
