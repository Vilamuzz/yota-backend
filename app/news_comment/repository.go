package news_comment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindOneComment(ctx context.Context, options map[string]interface{}) (*NewsComment, error)
	FindAllComments(ctx context.Context, options map[string]interface{}) ([]NewsComment, error)
	CountComments(ctx context.Context, options map[string]interface{}) (int64, error)
	CreateComment(ctx context.Context, comment *NewsComment) error
	UpdateComment(ctx context.Context, comment *NewsComment) error
	DeleteComment(ctx context.Context, commentID string) error
	FindReport(ctx context.Context, options map[string]interface{}) (*NewsCommentReport, error)
	CreateReport(ctx context.Context, report *NewsCommentReport) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindOneComment(ctx context.Context, options map[string]interface{}) (*NewsComment, error) {
	var comment NewsComment
	query := r.Conn.WithContext(ctx).
		Preload("Account.UserProfile").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Replies.Account.UserProfile")

	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		isReportSubquery := r.Conn.Table("news_comment_reports").
			Select("COUNT(*) > 0").
			Where("news_comment_id = news_comments.id AND account_id = ?", accountID.(string))
		query = query.Select("news_comments.*, (?) as is_reported", isReportSubquery)
		delete(options, "account_id")
	}

	if err := query.Where(options).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

var allowedCommentSortColumns = map[string]string{
	"created_at":   "created_at",
	"createdat":    "created_at",
	"report_count": "report_count",
	"reportcount":  "report_count",
}

func (r *repository) FindAllComments(ctx context.Context, options map[string]interface{}) ([]NewsComment, error) {
	var comments []NewsComment
	query := r.Conn.WithContext(ctx).Where("deleted_at IS NULL").
		Preload("Account.UserProfile").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Replies.Account.UserProfile")

	if topLevelOnly, ok := options["top_level_only"]; ok && topLevelOnly.(bool) {
		query = query.Where("parent_comment_id IS NULL")
	}

	if newsID, ok := options["news_id"]; ok {
		query = query.Where("news_id = ?", newsID)
	}
	if reported, ok := options["reported"]; ok {
		if reported.(bool) {
			query = query.Where("reported = ?", true)
		}
	}

	sortCol := "created_at"
	sortDir := "DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedCommentSortColumns[parts[0]]; valid {
				sortCol = col
				if len(parts) == 2 && (parts[1] == "asc" || parts[1] == "desc") {
					sortDir = strings.ToUpper(parts[1])
				}
			}
		}
	}

	query = query.Order(fmt.Sprintf("%s %s, id DESC", sortCol, sortDir))

	limit := 10
	if l, ok := options["limit"]; ok && l.(int) > 0 {
		limit = l.(int)
	}

	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *repository) CountComments(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&NewsComment{}).Where("deleted_at IS NULL")

	if topLevelOnly, ok := options["top_level_only"]; ok && topLevelOnly.(bool) {
		query = query.Where("parent_comment_id IS NULL")
	}
	if newsID, ok := options["news_id"]; ok {
		query = query.Where("news_id = ?", newsID)
	}
	if reported, ok := options["reported"]; ok {
		if reported.(bool) {
			query = query.Where("reported = ?", true)
		}
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *repository) CreateComment(ctx context.Context, comment *NewsComment) error {
	return r.Conn.WithContext(ctx).Create(comment).Error
}

func (r *repository) UpdateComment(ctx context.Context, comment *NewsComment) error {
	return r.Conn.WithContext(ctx).Save(comment).Error
}

func (r *repository) DeleteComment(ctx context.Context, commentID string) error {
	return r.Conn.WithContext(ctx).Model(&NewsComment{}).Where("id = ?", commentID).Update("deleted_at", time.Now()).Error
}

func (r *repository) FindReport(ctx context.Context, options map[string]interface{}) (*NewsCommentReport, error) {
	var report NewsCommentReport
	query := r.Conn.WithContext(ctx).Where(options).First(&report)
	if query.Error != nil {
		return nil, query.Error
	}
	return &report, nil
}

func (r *repository) CreateReport(ctx context.Context, report *NewsCommentReport) error {
	if err := r.Conn.WithContext(ctx).Create(report).Error; err != nil {
		return err
	}
	return r.Conn.WithContext(ctx).Model(&NewsComment{}).Where("id = ?", report.NewsCommentID).Updates(map[string]interface{}{
		"report_count": gorm.Expr("report_count + ?", 1),
		"reported":     true,
	}).Error
}
