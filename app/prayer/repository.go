package prayer

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	FindOnePrayer(ctx context.Context, options map[string]interface{}) (*Prayer, error)
	FindAllPrayers(ctx context.Context, options map[string]interface{}) ([]Prayer, error)
	CountPrayers(ctx context.Context, options map[string]interface{}) (int64, error)
	CreatePrayer(ctx context.Context, prayer *Prayer) error
	UpdatePrayer(ctx context.Context, prayer *Prayer) error
	DeletePrayer(ctx context.Context, prayerID string) error
	CreateAmen(ctx context.Context, amen *PrayerAmen) error
	DeleteAmen(ctx context.Context, prayerID, accountID string) (int64, error)
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
	query := r.Conn.WithContext(ctx).Preload("DonationProgramTransaction")

	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		isAmenSubquery := r.Conn.Table("prayer_amens").
			Select("COUNT(*) > 0").
			Where("prayer_id = prayers.id AND account_id = ?", accountID.(string))
		query = query.Select("prayers.*, (?) as is_amen", isAmenSubquery)
		delete(options, "account_id")
	}

	if _, ok := options["donation_program_transaction_id"]; !ok {
		query = query.Where("prayers.is_published = ?", true)
	}

	if err := query.Where(options).First(&prayer).Error; err != nil {
		return nil, err
	}
	return &prayer, nil
}

var allowedPrayerSortColumns = map[string]string{
	"created_at":   "created_at",
	"createdat":    "created_at",
	"report_count": "report_count",
	"reportcount":  "report_count",
	"amen_count":   "amen_count",
	"amencount":    "amen_count",
}

func (r *repository) FindAllPrayers(ctx context.Context, options map[string]interface{}) ([]Prayer, error) {
	var prayers []Prayer
	query := r.Conn.WithContext(ctx).Preload("DonationProgramTransaction").Where("prayers.is_published = ?", true)

	if donationProgramID, ok := options["donation_program_id"]; ok {
		query = query.Joins("JOIN donation_program_transactions ON donation_program_transactions.id = prayers.donation_program_transaction_id").
			Where("donation_program_transactions.donation_program_id = ?", donationProgramID)
	}

	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		isAmenSubquery := r.Conn.Table("prayer_amens").
			Select("COUNT(*) > 0").
			Where("prayer_id = prayers.id AND account_id = ?", accountID.(string))
		query = query.Select("prayers.*, (?) as is_amen", isAmenSubquery)
	}
	if donationID, ok := options["donation_id"]; ok {
		query = query.Where("donation_program_transaction_id = ?", donationID)
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
			if col, valid := allowedPrayerSortColumns[parts[0]]; valid {
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
	if err := query.Find(&prayers).Error; err != nil {
		return nil, err
	}
	return prayers, nil
}

func (r *repository) CountPrayers(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&Prayer{}).Where("prayers.is_published = ?", true)

	if donationProgramID, ok := options["donation_program_id"]; ok {
		query = query.Joins("JOIN donation_program_transactions ON donation_program_transactions.id = prayers.donation_program_transaction_id").
			Where("donation_program_transactions.donation_program_id = ?", donationProgramID)
	}
	if donationID, ok := options["donation_id"]; ok {
		query = query.Where("donation_program_transaction_id = ?", donationID)
	}
	if reported, ok := options["reported"]; ok {
		if reported.(bool) {
			query = query.Where("reported = ?", true)
		}
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *repository) CreatePrayer(ctx context.Context, prayer *Prayer) error {
	return r.Conn.WithContext(ctx).Create(prayer).Error
}

func (r *repository) UpdatePrayer(ctx context.Context, prayer *Prayer) error {
	return r.Conn.WithContext(ctx).Omit("DonationProgramTransaction").Save(prayer).Error
}

func (r *repository) DeletePrayer(ctx context.Context, prayerID string) error {
	tx := r.Conn.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("prayer_id = ?", prayerID).Delete(&PrayerAmen{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("prayer_id = ?", prayerID).Delete(&PrayerReport{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ?", prayerID).Delete(&Prayer{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *repository) CreateAmen(ctx context.Context, amen *PrayerAmen) error {
	if err := r.Conn.WithContext(ctx).Create(amen).Error; err != nil {
		return err
	}
	// Increment amen count
	return r.Conn.WithContext(ctx).Model(&Prayer{}).Where("id = ?", amen.PrayerID).Update("amen_count", gorm.Expr("amen_count + ?", 1)).Error
}

func (r *repository) DeleteAmen(ctx context.Context, prayerID, accountID string) (int64, error) {
	result := r.Conn.WithContext(ctx).Where("prayer_id = ? AND account_id = ?", prayerID, accountID).Delete(&PrayerAmen{})
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
	// Increment report count and set reported to true
	return r.Conn.WithContext(ctx).Model(&Prayer{}).Where("id = ?", report.PrayerID).Updates(map[string]interface{}{
		"report_count": gorm.Expr("report_count + ?", 1),
		"reported":     true,
	}).Error
}
