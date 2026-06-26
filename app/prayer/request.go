package prayer

type PrayerQueryParams struct {
	DonationSlug string `form:"-"`
	AccountID    string `form:"-"`
	Page         int    `form:"page"`
	Limit        int    `form:"limit"`
	SortBy       string `form:"sortBy"`
}