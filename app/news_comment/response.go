package news_comment

import (
	"github.com/Vilamuzz/yota-backend/pkg"
)

type NewsCommentResponse struct {
	ID        string                `json:"id"`
	Username  string                `json:"username"`
	Content   string                `json:"content"`
	Replies   []NewsCommentResponse `json:"replies"`
	CreatedAt string                `json:"createdAt"`
}

type AdminNewsCommentResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Content     string `json:"content"`
	CreatedAt   string `json:"createdAt"`
	ReportCount int64  `json:"reportCount"`
}

type AdminNewsCommentListResponse struct {
	Comments   []AdminNewsCommentResponse `json:"comments"`
	Pagination pkg.OffsetPagination       `json:"pagination"`
}

type NewsCommentListResponse struct {
	Comments   []NewsCommentResponse `json:"comments"`
	Pagination pkg.OffsetPagination  `json:"pagination"`
}

func (p *NewsComment) toNewsCommentResponse() NewsCommentResponse {
	username := p.Account.UserProfile.Username
	if username == "" {
		username = "Hamba Allah"
	}

	var replies []NewsCommentResponse
	for _, reply := range p.Replies {
		replies = append(replies, reply.toNewsCommentResponse())
	}
	if replies == nil {
		replies = []NewsCommentResponse{}
	}

	return NewsCommentResponse{
		ID:        p.ID.String(),
		Username:  username,
		Content:   p.Content,
		Replies:   replies,
		CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func toNewsCommentListResponse(comments []NewsComment, pagination pkg.OffsetPagination) NewsCommentListResponse {
	var responses []NewsCommentResponse
	for _, comment := range comments {
		responses = append(responses, comment.toNewsCommentResponse())
	}
	if responses == nil {
		responses = []NewsCommentResponse{}
	}
	return NewsCommentListResponse{
		Comments:   responses,
		Pagination: pagination,
	}
}

func (p *NewsComment) toAdminNewsCommentResponse() AdminNewsCommentResponse {
	username := p.Account.UserProfile.Username
	if username == "" {
		username = "Hamba Allah"
	}

	return AdminNewsCommentResponse{
		ID:          p.ID.String(),
		Username:    username,
		Content:     p.Content,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		ReportCount: p.ReportCount,
	}
}

func toAdminNewsCommentListResponse(comments []NewsComment, pagination pkg.OffsetPagination) AdminNewsCommentListResponse {
	var responses []AdminNewsCommentResponse
	for _, comment := range comments {
		responses = append(responses, comment.toAdminNewsCommentResponse())
	}
	if responses == nil {
		responses = []AdminNewsCommentResponse{}
	}
	return AdminNewsCommentListResponse{
		Comments:   responses,
		Pagination: pagination,
	}
}
