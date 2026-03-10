package donation

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindPublished(ctx context.Context, options map[string]interface{}) ([]Donation, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*Donation, error)
	FindActiveByID(ctx context.Context, id string) (*Donation, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]Donation, error)
	FindByID(ctx context.Context, id string) (*Donation, error)
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

func (r *repository) FindPublished(ctx context.Context, options map[string]interface{}) ([]Donation, error) {
	var donations []Donation
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donations.*, (?) as collected_fund", collectedFundSubquery)

	// Filter by active status
	query = query.Where("status != ?", StatusDraft)

	// Apply filters
	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}

	// Apply cursor-based pagination
	if cursor, ok := options["cursor"]; ok && cursor != "" {
		cursorStr := cursor.(string)
		cursorData, err := pkg.DecodeCursor(cursorStr)
		if err == nil {
			// Cursor format: created_at|id
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	// Apply limit (fetch one extra to check if there's a next page)
	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	query = query.Limit(limit + 1)

	// Order by created date (newest first) and ID for consistent ordering
	query = query.Order("created_at DESC, id DESC")

	if err := query.Find(&donations).Error; err != nil {
		return nil, err
	}

	return donations, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Donation, error) {
	var donations []Donation
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donations.*, (?) as collected_fund", collectedFundSubquery)

	// Apply filters
	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Apply cursor-based pagination
	if cursor, ok := options["cursor"]; ok && cursor != "" {
		cursorStr := cursor.(string)
		cursorData, err := pkg.DecodeCursor(cursorStr)
		if err == nil {
			// Cursor format: created_at|id
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	}

	// Apply limit (fetch one extra to check if there's a next page)
	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	query = query.Limit(limit + 1)

	// Order by created date (newest first) and ID for consistent ordering
	query = query.Order("created_at DESC, id DESC")

	if err := query.Find(&donations).Error; err != nil {
		return nil, err
	}

	return donations, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (*Donation, error) {
	var donation Donation
	if err := r.Conn.WithContext(ctx).
		Select("donations.*").
		Where("id = ?", id).
		First(&donation).Error; err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *repository) FindPublishedBySlug(ctx context.Context, slug string) (*Donation, error) {
	var donation Donation
	collectedFundSubquery := r.Conn.Table("donation_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donations.id AND transaction_status = 'settlement'")
	if err := r.Conn.WithContext(ctx).
		Select("donations.*, (?) as collected_fund", collectedFundSubquery).
		Where("slug = ? AND status != ?", slug, StatusDraft).
		First(&donation).Error; err != nil {
		return nil, err
	}
	return &donation, nil
}

func (r *repository) FindActiveByID(ctx context.Context, id string) (*Donation, error) {
	var donation Donation
	if err := r.Conn.WithContext(ctx).
		Where("id = ? AND status = ?", id, StatusActive).
		First(&donation).Error; err != nil {
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
