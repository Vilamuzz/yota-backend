package donation_program

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllDonationPrograms(ctx context.Context, options map[string]interface{}) ([]DonationProgram, error)
	CountDonationPrograms(ctx context.Context, options map[string]interface{}) (int64, error)
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

// allowedSortColumns whitelists sortable columns to prevent SQL injection.
var allowedSortColumns = map[string]string{
	"title":          "dp.title",
	"fund_target":    "dp.fund_target",
	"collected_fund": "COALESCE(dpt.collected_fund, 0)",
	"total_expense":  "COALESCE(dpe.total_expense, 0)",
	"start_date":     "dp.start_date",
	"end_date":       "dp.end_date",
	"created_at":     "dp.created_at",
	"status":         "dp.status",
}

func buildDonationProgramBaseQuery(conn *gorm.DB, ctx context.Context, options map[string]interface{}) *gorm.DB {
	dptSubquery := conn.Table("donation_program_transactions").
		Select("donation_program_id, COALESCE(SUM(gross_amount), 0) as collected_fund").
		Where("transaction_status = 'settlement'").
		Group("donation_program_id")

	dpeSubquery := conn.Table("donation_program_expenses").
		Select("donation_program_id, COALESCE(SUM(amount), 0) as total_expense").
		Where("deleted_at IS NULL").
		Group("donation_program_id")

	query := conn.WithContext(ctx).
		Table("donation_programs dp").
		Joins("LEFT JOIN (?) dpt ON dpt.donation_program_id = dp.id", dptSubquery).
		Joins("LEFT JOIN (?) dpe ON dpe.donation_program_id = dp.id", dpeSubquery).
		Where("dp.deleted_at IS NULL").
		Select("dp.*, COALESCE(dpt.collected_fund, 0) as collected_fund, COALESCE(dpe.total_expense, 0) as total_expense")

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("dp.title ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("dp.category = ?", category)
	}
	if status, ok := options["status"]; ok {
		switch v := status.(type) {
		case Status:
			if v != "" {
				query = query.Where("dp.status = ?", string(v))
			}
		case []Status:
			if len(v) > 0 {
				query = query.Where("dp.status IN ?", v)
			}
		}
	}
	return query
}

func (r *repository) FindAllDonationPrograms(ctx context.Context, options map[string]interface{}) ([]DonationProgram, error) {
	var donationPrograms []DonationProgram
	query := buildDonationProgramBaseQuery(r.Conn, ctx, options)

	orderClause := "dp.created_at DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedSortColumns[parts[0]]; valid {
				dir := "ASC"
				if len(parts) == 2 && parts[1] == "desc" {
					dir = "DESC"
				}
				orderClause = fmt.Sprintf("%s %s", col, dir)
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
	if err := query.Find(&donationPrograms).Error; err != nil {
		return nil, err
	}
	return donationPrograms, nil
}

func (r *repository) CountDonationPrograms(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&DonationProgram{}).Where("deleted_at IS NULL")
	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if status, ok := options["status"]; ok {
		switch v := status.(type) {
		case Status:
			if v != "" {
				query = query.Where("status = ?", string(v))
			}
		case []Status:
			if len(v) > 0 {
				query = query.Where("status IN ?", v)
			}
		}
	}
	err := query.Count(&total).Error
	return total, err
}

func (r *repository) FindOneDonationProgram(ctx context.Context, options map[string]interface{}) (*DonationProgram, error) {
	var donationProgram DonationProgram
	dptSubquery := r.Conn.Table("donation_program_transactions").
		Select("donation_program_id, COALESCE(SUM(gross_amount), 0) as collected_fund").
		Where("transaction_status = 'settlement'").
		Group("donation_program_id")

	dpeSubquery := r.Conn.Table("donation_program_expenses").
		Select("donation_program_id, COALESCE(SUM(amount), 0) as total_expense").
		Where("deleted_at IS NULL").
		Group("donation_program_id")

	query := r.Conn.WithContext(ctx).
		Table("donation_programs dp").
		Joins("LEFT JOIN (?) dpt ON dpt.donation_program_id = dp.id", dptSubquery).
		Joins("LEFT JOIN (?) dpe ON dpe.donation_program_id = dp.id", dpeSubquery).
		Where("dp.deleted_at IS NULL").
		Select("dp.*, COALESCE(dpt.collected_fund, 0) as collected_fund, COALESCE(dpe.total_expense, 0) as total_expense")

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("dp.id = ?", id)
	}

	if title, ok := options["title"]; ok && title != "" {
		query = query.Where("dp.title = ?", title)
	}

	if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("dp.slug = ?", slug)
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
		Where("donation_program_id = donation_programs.id AND transaction_status = 'settlement'")

	return r.Conn.WithContext(ctx).
		Model(&DonationProgram{}).
		Where("end_date < NOW() AND status = ? AND deleted_at IS NULL", StatusActive).
		Update("status", gorm.Expr("CASE WHEN (?) >= fund_target THEN ? ELSE ? END",
			collectedFundSubquery, StatusCompleted, StatusExpired)).Error
}
