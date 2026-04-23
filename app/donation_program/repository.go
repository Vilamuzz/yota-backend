package donation_program

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllDonationPrograms(ctx context.Context, options map[string]interface{}) ([]DonationProgram, error)
	FindOneDonationProgram(ctx context.Context, options map[string]interface{}) (*DonationProgram, error)
	CreateDonationProgram(ctx context.Context, donationProgram *DonationProgram) error
	UpdateDonationProgram(ctx context.Context, donationProgramID string, updateData map[string]interface{}) error
	DeleteDonationProgram(ctx context.Context, donationProgramID string) error
	UpdateExpiredDonationProgram(ctx context.Context) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FindAllDonationPrograms(ctx context.Context, options map[string]interface{}) ([]DonationProgram, error) {
	var donationPrograms []DonationProgram
	collectedFundSubquery := r.Conn.Table("donation_program_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_program_id = donation_programs.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donation_programs.*, (?) as collected_fund", collectedFundSubquery).
		Where("deleted_at IS NULL")

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
	if err := query.Find(&donationPrograms).Error; err != nil {
		return nil, err
	}
	return donationPrograms, nil
}

func (r *repository) FindOneDonationProgram(ctx context.Context, options map[string]interface{}) (*DonationProgram, error) {
	var donationProgram DonationProgram
	collectedFundSubquery := r.Conn.Table("donation_program_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_program_id = donation_programs.id AND transaction_status = 'settlement'")
	query := r.Conn.WithContext(ctx).
		Select("donation_programs.*, (?) as collected_fund", collectedFundSubquery).
		Where("deleted_at IS NULL")
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
	if err := query.First(&donationProgram).Error; err != nil {
		return nil, err
	}
	return &donationProgram, nil
}

func (r *repository) CreateDonationProgram(ctx context.Context, donationProgram *DonationProgram) error {
	return r.Conn.WithContext(ctx).Create(donationProgram).Error
}

func (r *repository) UpdateDonationProgram(ctx context.Context, donationProgramID string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&DonationProgram{}).Where("id = ?", donationProgramID).Updates(updateData).Error
}

func (r *repository) DeleteDonationProgram(ctx context.Context, donationProgramID string) error {
	return r.Conn.WithContext(ctx).Model(&DonationProgram{}).Where("id = ?", donationProgramID).Update("deleted_at", time.Now()).Error
}

func (r *repository) UpdateExpiredDonationProgram(ctx context.Context) error {
	collectedFundSubquery := r.Conn.Table("donation_program_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("donation_id = donation_programs.id AND transaction_status = 'settlement'")

	return r.Conn.WithContext(ctx).
		Model(&DonationProgram{}).
		Where("end_date < NOW() AND status = ? AND deleted_at IS NULL", StatusActive).
		Update("status", gorm.Expr("CASE WHEN (?) >= fund_target THEN ? ELSE ? END",
			collectedFundSubquery, StatusCompleted, StatusExpired)).Error
}
