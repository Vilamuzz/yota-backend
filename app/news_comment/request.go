package news_comment

type NewsCommentQueryParams struct {
	NewsSlug  string `form:"-"`
	AccountID string `form:"-"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
	SortBy    string `form:"sortBy"`
}

type CreateNewsCommentRequest struct {
	ParentCommentID *string `json:"parentCommentId"`
	Content         string  `json:"content"`
}

type ReportNewsCommentRequest struct {
}
