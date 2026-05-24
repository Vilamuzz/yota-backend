package social_program

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialPrograms(ctx context.Context, options map[string]interface{}) ([]SocialProgram, error)
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

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("status = ?", status.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	if _, isPrev := options["prev_cursor"]; isPrev {
		query = query.Order("created_at ASC, id ASC")
	} else {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&socialPrograms).Error; err != nil {
		return nil, err
	}
	return socialPrograms, nil
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
	return r.Conn.WithContext(ctx).Where("id = ?", socialProgramID).Update("deleted_at", time.Now()).Error
}
