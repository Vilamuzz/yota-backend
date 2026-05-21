package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerQueryParams struct {
	pkg.PaginationParams
}

type ReportPrayerRequest struct {
	Reason string `json:"reason"`
}
