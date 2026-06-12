package news_comment

import "github.com/Vilamuzz/yota-backend/pkg"

type NewsCommentQueryParams struct {
	pkg.PaginationParams
}

type CreateNewsCommentRequest struct {
	ParentCommentID *string `json:"parentCommentId"`
	Content         string  `json:"content"`
}

type ReportNewsCommentRequest struct {
	Reason string `json:"reason"`
}
