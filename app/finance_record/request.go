package finance_record

type RecordQueryParams struct {
	FundID     string `form:"fund_id"`
	SourceType string `form:"source_type"`
	NextCursor string `form:"next_cursor"`
	PrevCursor string `form:"prev_cursor"`
	Limit      int    `form:"limit"`
}
