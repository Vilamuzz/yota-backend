package donation

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context, options map[string]interface{}) ([]Donation, error)
	FindOne(ctx context.Context, options map[string]interface{}) (*Donation, error)
	Create(ctx context.Context, donation *Donation) error
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	UpdateExpired(ctx context.Context) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Donation, error) {
	var donations []Donation
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donations.*, (?) as collected_fund", collectedFundSubquery)

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if published, ok := options["published"]; ok && published == true {
		query = query.Where("status != ?", StatusDraft)
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID).
				Order("created_at ASC, id ASC")
		}
	}

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&donations).Error; err != nil {
		return nil, err
	}
	return donations, nil
}

func (r *repository) FindOne(ctx context.Context, options map[string]interface{}) (*Donation, error) {
	var donation Donation
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donations.*, (?) as collected_fund", collectedFundSubquery)
	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}
	if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("slug = ?", slug)
	}
	if published, ok := options["published"]; ok && published == true {
		query = query.Where("status != ?", StatusDraft)
	}
	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.First(&donation).Error; err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *repository) Create(ctx context.Context, donation *Donation) error {
	return r.Conn.WithContext(ctx).Create(donation).Error
}

func (r *repository) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Donation{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Model(&Donation{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *repository) UpdateExpired(ctx context.Context) error {
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")

	return r.Conn.WithContext(ctx).
		Model(&Donation{}).
		Where("date_end < NOW() AND status = ?", StatusActive).
		Update("status", gorm.Expr("CASE WHEN (?) >= fund_target THEN ? ELSE ? END",
			collectedFundSubquery, StatusCompleted, StatusExpired)).Error
}
