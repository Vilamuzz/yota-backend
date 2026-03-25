package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerQueryParams struct {
	DonationID string `form:"donation_id"`
	Reported   bool   `form:"reported"`
	pkg.PaginationParams
}

type ReportPrayerRequest struct {
	Reason string `json:"reason"`
}
