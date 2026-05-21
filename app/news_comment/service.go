package news_comment

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/app/news"
	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	GetNewsCommentList(ctx context.Context, newsID, accountID string, isAdmin bool, params NewsCommentQueryParams) pkg.Response
	GetNewsCommentByID(ctx context.Context, newsCommentID, accountID string) pkg.Response
	CreateNewsComment(ctx context.Context, newsSlug, accountID string, payload CreateNewsCommentRequest) pkg.Response
	DeleteNewsComment(ctx context.Context, newsCommentID string) pkg.Response
	CreateReportNewsComment(ctx context.Context, newsCommentID, accountID string, payload ReportNewsCommentRequest) pkg.Response
	AllowNewsComment(ctx context.Context, newsCommentID string) pkg.Response
}

type service struct {
	repo     Repository
	newsRepo news.Repository
	timeout  time.Duration
}

func NewService(repo Repository, newsRepo news.Repository, timeout time.Duration) Service {
	return &service{repo: repo, newsRepo: newsRepo, timeout: timeout}
}

func (s *service) CreateReportNewsComment(ctx context.Context, newsCommentID, accountID string, payload ReportNewsCommentRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(newsCommentID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID komentar tidak valid"}, nil)
	}
	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}
	newsComment, err := s.repo.FindOneComment(ctx, map[string]interface{}{"id": newsCommentID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Komentar tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
	}

	if newsComment.Reported != nil && !*newsComment.Reported {
		return pkg.NewResponse(http.StatusOK, "Komentar berhasil dilaporkan", nil, nil)
	}

	if payload.Reason == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan Validasi", map[string]string{
			"reason": "Alasan wajib diisi",
		}, nil)
	}

	_, err = s.repo.FindReport(ctx, map[string]interface{}{
		"news_comment_id": newsCommentID,
		"account_id":      accountID,
	})
	if err == nil {
		return pkg.NewResponse(http.StatusOK, "Komentar berhasil dilaporkan", nil, nil)
	}

	report := &NewsCommentReport{
		NewsCommentID: uuid.MustParse(newsCommentID),
		AccountID:     uuid.MustParse(accountID),
		Reason:        payload.Reason,
	}
	if err := s.repo.CreateReport(ctx, report); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":       "news_comment.service",
			"news_comment_id": newsCommentID,
			"account_id":      accountID,
		}).WithError(err).Error("failed to create news comment report")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal melaporkan komentar", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Komentar berhasil dilaporkan", nil, nil)
}

func (s *service) AllowNewsComment(ctx context.Context, newsCommentID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	if err := uuid.Validate(newsCommentID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID komentar tidak valid"}, nil)
	}

	newsComment, err := s.repo.FindOneComment(ctx, map[string]interface{}{"id": newsCommentID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Komentar tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
	}
	reported := false
	newsComment.Reported = &reported
	if err := s.repo.UpdateComment(ctx, newsComment); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":       "news_comment.service",
			"news_comment_id": newsCommentID,
		}).WithError(err).Error("failed to update news comment")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui komentar", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Komentar berhasil diizinkan", nil, nil)
}

func (s *service) CreateNewsComment(ctx context.Context, newsSlug, accountID string, payload CreateNewsCommentRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(accountID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"account_id": "Format ID akun tidak valid"}, nil)
	}

	news, err := s.newsRepo.FindOneNews(ctx, map[string]interface{}{"slug": newsSlug})
	if err != nil {
		return pkg.NewResponse(http.StatusNotFound, "Berita tidak ditemukan", nil, nil)
	}

	if payload.Content == "" {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"content": "Konten tidak boleh kosong"}, nil)
	}

	newComment := &NewsComment{
		ID:        uuid.New(),
		NewsID:    news.ID,
		AccountID: uuid.MustParse(accountID),
		Content:   payload.Content,
		CreatedAt: time.Now(),
	}

	if payload.ParentCommentID != nil && *payload.ParentCommentID != "" {
		if err := uuid.Validate(*payload.ParentCommentID); err != nil {
			return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"parentCommentId": "Format ID komentar tidak valid"}, nil)
		}
		
		parentComment, err := s.repo.FindOneComment(ctx, map[string]interface{}{"id": *payload.ParentCommentID})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return pkg.NewResponse(http.StatusNotFound, "Komentar tidak ditemukan", nil, nil)
			}
			return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
		}

		if parentComment.ParentCommentID != nil {
			return pkg.NewResponse(http.StatusBadRequest, "Tidak dapat membalas balasan komentar", nil, nil)
		}
		
		parentID := uuid.MustParse(*payload.ParentCommentID)
		newComment.ParentCommentID = &parentID
	}

	if err := s.repo.CreateComment(ctx, newComment); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "news_comment.service",
			"news_slug": newsSlug,
			"account_id": accountID,
		}).WithError(err).Error("failed to create news comment")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menambahkan komentar", nil, nil)
	}

	// Fetch with profile
	createdComment, _ := s.repo.FindOneComment(ctx, map[string]interface{}{"id": newComment.ID.String()})
	if createdComment != nil {
		return pkg.NewResponse(http.StatusOK, "Berhasil menambahkan komentar", nil, createdComment.toNewsCommentResponse())
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil menambahkan komentar", nil, newComment.toNewsCommentResponse())
}

func (s *service) GetNewsCommentByID(ctx context.Context, newsCommentID, accountID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(newsCommentID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID komentar tidak valid"}, nil)
	}
	newsComment, err := s.repo.FindOneComment(ctx, map[string]interface{}{"id": newsCommentID, "account_id": accountID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Komentar tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan komentar", nil, newsComment.toNewsCommentResponse())
}

func (s *service) GetNewsCommentList(ctx context.Context, accountID, newsSlug string, isAdmin bool, params NewsCommentQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit == 0 {
		params.Limit = 10
	}

	options := map[string]interface{}{
		"limit":          params.Limit,
		"account_id":     accountID,
		"top_level_only": true,
	}

	if isAdmin {
		options["reported"] = true
	}

	pagination := pkg.CursorPagination{
		Limit: params.Limit,
	}

	if newsSlug != "" {
		news, err := s.newsRepo.FindOneNews(ctx, map[string]interface{}{"slug": newsSlug})
		if err != nil {
			return pkg.NewResponse(http.StatusOK, "Berhasil menemukan komentar", nil, toNewsCommentListResponse([]NewsComment{}, pagination))
		}
		options["news_id"] = news.ID.String()
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}
	if params.PrevCursor != "" {
		options["prev_cursor"] = params.PrevCursor
	}

	newsComments, err := s.repo.FindAllComments(ctx, options)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
	}

	var hasNext, hasPrev bool
	if params.PrevCursor != "" {
		hasPrev = len(newsComments) > params.Limit
		hasNext = true
		if len(newsComments) > params.Limit {
			newsComments = newsComments[:params.Limit]
		}
		for i, j := 0, len(newsComments)-1; i < j; i, j = i+1, j-1 {
			newsComments[i], newsComments[j] = newsComments[j], newsComments[i]
		}
	} else {
		hasNext = len(newsComments) > params.Limit
		hasPrev = params.NextCursor != ""
		if hasNext {
			newsComments = newsComments[:params.Limit]
		}
	}

	var nextCursor, prevCursor string
	if len(newsComments) > 0 {
		first := newsComments[0]
		last := newsComments[len(newsComments)-1]
		if hasNext {
			nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID.String())
		}
		if hasPrev {
			prevCursor = pkg.EncodeCursor(first.CreatedAt, first.ID.String())
		}
	}

	pagination = pkg.CursorPagination{
		NextCursor: nextCursor,
		PrevCursor: prevCursor,
		Limit:      params.Limit,
	}

	if isAdmin {
		return pkg.NewResponse(http.StatusOK, "Berhasil menemukan komentar", nil, toAdminNewsCommentListResponse(newsComments, pagination))
	}
	return pkg.NewResponse(http.StatusOK, "Berhasil menemukan komentar", nil, toNewsCommentListResponse(newsComments, pagination))
}

func (s *service) DeleteNewsComment(ctx context.Context, newsCommentID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if err := uuid.Validate(newsCommentID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Kesalahan validasi", map[string]string{"id": "Format ID komentar tidak valid"}, nil)
	}

	_, err := s.repo.FindOneComment(ctx, map[string]interface{}{"id": newsCommentID})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return pkg.NewResponse(http.StatusNotFound, "Komentar tidak ditemukan", nil, nil)
		}
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan komentar", nil, nil)
	}

	if err := s.repo.DeleteComment(ctx, newsCommentID); err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus komentar", nil, nil)
	}
	return pkg.NewResponse(http.StatusOK, "Komentar berhasil dihapus", nil, nil)
}
