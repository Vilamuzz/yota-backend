package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerQueryParams struct {
	DonationID string `form:"donation_id"`
	Reported   bool   `form:"reported"`
	pkg.PaginationParams
}

type PrayerAmenRequest struct {
	PrayerID string `json:"prayer_id"`
}

type ReportPrayerRequest struct {
	PrayerID string `json:"prayer_id"`
	Reason   string `json:"reason"`
}
