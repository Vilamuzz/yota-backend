package social_program

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialPrograms(ctx context.Context, options map[string]interface{}) ([]SocialProgram, error)
	CountSocialPrograms(ctx context.Context, options map[string]interface{}) (int64, error)
	FindOneSocialProgram(ctx context.Context, options map[string]interface{}) (*SocialProgram, error)
	CreateSocialProgram(ctx context.Context, socialProgram *SocialProgram) error
	UpdateSocialProgram(ctx context.Context, socialProgramID string, updates map[string]interface{}) error
	DeleteSocialProgram(ctx context.Context, socialProgramID string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

// allowedSocialProgramSortColumns whitelists sortable columns to prevent SQL injection.
var allowedSocialProgramSortColumns = map[string]string{
	"title":             "title",
	"minimum_amount":    "minimum_amount",
	"billing_day":       "billing_day",
	"status":            "status",
	"created_at":        "created_at",
	"total_subscribers": "total_subscribers",
}

func (r *repository) FindAllSocialPrograms(ctx context.Context, options map[string]interface{}) ([]SocialProgram, error) {
	var socialPrograms []SocialProgram
	subscribersSubquery := r.Conn.Table("social_program_subscriptions").
		Select("COUNT(*)").
		Where("social_program_id = social_programs.id AND status = 'active'")

	collectedFundSubquery := r.Conn.Table("social_program_transactions spt").
		Select("COALESCE(SUM(spt.gross_amount), 0)").
		Joins("JOIN social_program_invoices spi ON spt.social_program_invoice_id = spi.id").
		Joins("JOIN social_program_subscriptions sps ON spi.subscription_id = sps.id").
		Where("sps.social_program_id = social_programs.id AND spt.transaction_status = 'settlement'")

	totalExpenseSubquery := r.Conn.Table("social_program_expenses").
		Select("COALESCE(SUM(amount), 0)").
		Where("social_program_id = social_programs.id")

	query := r.Conn.WithContext(ctx).
		Select("social_programs.*, (?) as total_subscribers, (?) as collected_fund, (?) as total_expense", subscribersSubquery, collectedFundSubquery, totalExpenseSubquery).
		Where("deleted_at IS NULL")

	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		isSubscribedSubquery := r.Conn.Table("social_program_subscriptions").
			Select("COUNT(*) > 0").
			Where("social_program_id = social_programs.id AND account_id = ? AND status = 'active'", accountID.(string))
		subscriptionIDSubquery := r.Conn.Table("social_program_subscriptions").
			Select("id").
			Where("social_program_id = social_programs.id AND account_id = ? AND status = 'active'", accountID.(string)).
			Limit(1)
		query = query.Select("social_programs.*, (?) as total_subscribers, (?) as collected_fund, (?) as total_expense, (?) as is_subscribed, (?) as subscription_id", subscribersSubquery, collectedFundSubquery, totalExpenseSubquery, isSubscribedSubquery, subscriptionIDSubquery)
	}

	if status, ok := options["status"]; ok {
		switch v := status.(type) {
		case string:
			if v != "" {
				query = query.Where("status = ?", v)
			}
		case []string:
			if len(v) > 0 {
				query = query.Where("status IN ?", v)
			}
		}
	}
	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if startDay, ok := options["start_day"]; ok {
		if endDay, ok := options["end_day"]; ok {
			sDay := startDay.(int)
			eDay := endDay.(int)
			if sDay <= eDay {
				query = query.Where("billing_day >= ? AND billing_day <= ?", sDay, eDay)
			} else {
				query = query.Where("billing_day >= ? OR billing_day <= ?", sDay, eDay)
			}
		}
	}

	// Build ORDER BY from "sort_by" option, e.g. "title asc" or "minimum_amount desc".
	orderClause := "social_programs.created_at DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedSocialProgramSortColumns[parts[0]]; valid {
				dir := "ASC"
				if len(parts) == 2 && parts[1] == "desc" {
					dir = "DESC"
				}
				if col == "total_subscribers" {
					orderClause = fmt.Sprintf("(SELECT COUNT(*) FROM social_program_subscriptions WHERE social_program_id = social_programs.id AND status = 'active') %s", dir)
				} else {
					orderClause = fmt.Sprintf("social_programs.%s %s", col, dir)
				}
			}
		}
	}
	query = query.Order(orderClause)

	limit := 10
	if l, ok := options["limit"]; ok && l.(int) > 0 {
		limit = l.(int)
	}
	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&socialPrograms).Error; err != nil {
		return nil, err
	}
	return socialPrograms, nil
}

func (r *repository) CountSocialPrograms(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&SocialProgram{}).Where("deleted_at IS NULL")
	if status, ok := options["status"]; ok {
		switch v := status.(type) {
		case string:
			if v != "" {
				query = query.Where("status = ?", v)
			}
		case []string:
			if len(v) > 0 {
				query = query.Where("status IN ?", v)
			}
		}
	}
	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if startDay, ok := options["start_day"]; ok {
		if endDay, ok := options["end_day"]; ok {
			sDay := startDay.(int)
			eDay := endDay.(int)
			if sDay <= eDay {
				query = query.Where("billing_day >= ? AND billing_day <= ?", sDay, eDay)
			} else {
				query = query.Where("billing_day >= ? OR billing_day <= ?", sDay, eDay)
			}
		}
	}
	err := query.Count(&total).Error
	return total, err
}

func (r *repository) FindOneSocialProgram(ctx context.Context, options map[string]interface{}) (*SocialProgram, error) {
	var socialProgram SocialProgram
	subscribersSubquery := r.Conn.Table("social_program_subscriptions").
		Select("COUNT(*)").
		Where("social_program_id = social_programs.id AND status = 'active'")

	collectedFundSubquery := r.Conn.Table("social_program_transactions spt").
		Select("COALESCE(SUM(spt.gross_amount), 0)").
		Joins("JOIN social_program_invoices spi ON spt.social_program_invoice_id = spi.id").
		Joins("JOIN social_program_subscriptions sps ON spi.subscription_id = sps.id").
		Where("sps.social_program_id = social_programs.id AND spt.transaction_status = 'settlement'")

	totalExpenseSubquery := r.Conn.Table("social_program_expenses").
		Select("COALESCE(SUM(amount), 0)").
		Where("social_program_id = social_programs.id")

	query := r.Conn.WithContext(ctx).
		Select("social_programs.*, (?) as total_subscribers, (?) as collected_fund, (?) as total_expense", subscribersSubquery, collectedFundSubquery, totalExpenseSubquery).
		Where("deleted_at IS NULL")

	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		isSubscribedSubquery := r.Conn.Table("social_program_subscriptions").
			Select("COUNT(*) > 0").
			Where("social_program_id = social_programs.id AND account_id = ? AND status = 'active'", accountID.(string))
		subscriptionIDSubquery := r.Conn.Table("social_program_subscriptions").
			Select("id").
			Where("social_program_id = social_programs.id AND account_id = ? AND status = 'active'", accountID.(string)).
			Limit(1)
		query = query.Select("social_programs.*, (?) as total_subscribers, (?) as collected_fund, (?) as total_expense, (?) as is_subscribed, (?) as subscription_id", subscribersSubquery, collectedFundSubquery, totalExpenseSubquery, isSubscribedSubquery, subscriptionIDSubquery)
	}

	if id, ok := options["id"]; ok && id.(string) != "" {
		query = query.Where("id = ?", id.(string))
	}
	if slug, ok := options["slug"]; ok && slug.(string) != "" {
		query = query.Where("slug = ?", slug.(string))
	}
	if title, ok := options["title"]; ok && title.(string) != "" {
		query = query.Where("title = ?", title.(string))
	}

	if err := query.First(&socialProgram).Error; err != nil {
		return nil, err
	}
	return &socialProgram, nil
}

func (r *repository) CreateSocialProgram(ctx context.Context, socialProgram *SocialProgram) error {
	return r.Conn.WithContext(ctx).Create(socialProgram).Error
}

func (r *repository) UpdateSocialProgram(ctx context.Context, socialProgramID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgram{}).
		Where("id = ?", socialProgramID).
		Updates(updates).Error
}

func (r *repository) DeleteSocialProgram(ctx context.Context, socialProgramID string) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgram{}).Where("id = ?", socialProgramID).Update("deleted_at", time.Now()).Error
}
