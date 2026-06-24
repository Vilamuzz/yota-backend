package finance_record

import "github.com/Vilamuzz/yota-backend/pkg"

type RecordQueryParams struct {
	FundID     string `form:"fund_id"`
	SourceType string `form:"source_type"`
	pkg.PaginationParams
}

type MonthlyTrendQueryParams struct {
	Module string `form:"module"`
	Year   int    `form:"year"`
}
