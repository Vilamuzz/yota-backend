package prayer

import "github.com/Vilamuzz/yota-backend/pkg"

type PrayerQueryParams struct {
	Reported bool `form:"reported"`
	pkg.PaginationParams
}

type ReportPrayerRequest struct {
	Reason string `json:"reason"`
}
