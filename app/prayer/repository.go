package prayer

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindOnePrayer(ctx context.Context, options map[string]interface{}) (*Prayer, error)
	FindAllPrayers(ctx context.Context, options map[string]interface{}) ([]Prayer, error)
	CreatePrayer(ctx context.Context, prayer *Prayer) error
	UpdatePrayer(ctx context.Context, prayer *Prayer) error
	DeletePrayer(ctx context.Context, prayerID string) error
	FindAmenPrayerIDs(ctx context.Context, accountID string, prayerIDs []string) (map[string]bool, error)
	CreateAmen(ctx context.Context, amen *PrayerAmen) error
	DeleteAmen(ctx context.Context, prayerID, accountID string) (int64, error)
	ExistsAmen(ctx context.Context, prayerID, accountID string) (bool, error)
	FindReport(ctx context.Context, options map[string]interface{}) (*PrayerReport, error)
	CreateReport(ctx context.Context, report *PrayerReport) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindOnePrayer(ctx context.Context, options map[string]interface{}) (*Prayer, error) {
	var prayer Prayer
	query := r.Conn.WithContext(ctx).Preload("DonationProgramTransaction").Where(options)
	if err := query.First(&prayer).Error; err != nil {
		return nil, err
	}
	return &prayer, nil
}

func (r *repository) FindAllPrayers(ctx context.Context, options map[string]interface{}) ([]Prayer, error) {
	var prayers []Prayer
	query := r.Conn.WithContext(ctx).Preload("DonationProgramTransaction")

	if donationProgramID, ok := options["donation_program_id"]; ok {
		query = query.Joins("JOIN donation_program_transactions ON donation_program_transactions.id = prayers.donation_transaction_id").
			Where("donation_program_transactions.donation_program_id = ?", donationProgramID)
	}
	if donationID, ok := options["donation_id"]; ok {
		query = query.Where("donation_transaction_id = ?", donationID)
	}
	if reported, ok := options["reported"]; ok {
		if reported.(bool) {
			query = query.Where("report_count > ?", 0)
		}
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
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
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
	if err := query.Find(&prayers).Error; err != nil {
		return nil, err
	}
	return prayers, nil
}

func (r *repository) CreatePrayer(ctx context.Context, prayer *Prayer) error {
	return r.Conn.WithContext(ctx).Create(prayer).Error
}

func (r *repository) UpdatePrayer(ctx context.Context, prayer *Prayer) error {
	return r.Conn.WithContext(ctx).Save(prayer).Error
}

func (r *repository) DeletePrayer(ctx context.Context, prayerID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", prayerID).Update("deleted_at", time.Now()).Error
}

func (r *repository) FindAmenPrayerIDs(ctx context.Context, accountID string, prayerIDs []string) (map[string]bool, error) {
	amenMap := make(map[string]bool)
	if len(prayerIDs) == 0 {
		return amenMap, nil
	}

	var records []PrayerAmen
	err := r.Conn.WithContext(ctx).
		Model(&PrayerAmen{}).
		Select("prayer_id").
		Where("user_id = ? AND prayer_id IN ?", accountID, prayerIDs).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return amenMap, nil
}

func (r *repository) CreateAmen(ctx context.Context, amen *PrayerAmen) error {
	if err := r.Conn.WithContext(ctx).Create(amen).Error; err != nil {
		return err
	}
	// Increment amen count
	return r.Conn.WithContext(ctx).Model(&Prayer{}).Where("id = ?", amen.PrayerID).Update("amen_count", gorm.Expr("amen_count + ?", 1)).Error
}

func (r *repository) DeleteAmen(ctx context.Context, prayerID, accountID string) (int64, error) {
	result := r.Conn.WithContext(ctx).Where("prayer_id = ? AND user_id = ?", prayerID, accountID).Delete(&PrayerAmen{})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected > 0 {
		// Decrement amen count only if a row was actually deleted
		err := r.Conn.WithContext(ctx).Model(&Prayer{}).Where("id = ?", prayerID).Update("amen_count", gorm.Expr("GREATEST(amen_count - ?, 0)", result.RowsAffected)).Error
		return result.RowsAffected, err
	}
	return 0, nil
}

func (r *repository) ExistsAmen(ctx context.Context, prayerID, accountID string) (bool, error) {
	var count int64
	err := r.Conn.WithContext(ctx).
		Model(&PrayerAmen{}).
		Where("prayer_id = ? AND user_id = ?", prayerID, accountID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) FindReport(ctx context.Context, options map[string]interface{}) (*PrayerReport, error) {
	var report PrayerReport
	query := r.Conn.WithContext(ctx).Where(options).First(&report)
	if query.Error != nil {
		return nil, query.Error
	}
	return &report, nil
}

func (r *repository) CreateReport(ctx context.Context, report *PrayerReport) error {
	if err := r.Conn.WithContext(ctx).Create(report).Error; err != nil {
		return err
	}
	// Increment report count
	return r.Conn.WithContext(ctx).Model(&Prayer{}).Where("id = ?", report.PrayerID).Update("report_count", gorm.Expr("report_count + ?", 1)).Error
}
