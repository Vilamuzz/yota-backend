package finance_record

import (
	"context"
	"fmt"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"

	"github.com/Vilamuzz/yota-backend/app/donation_program"
	"github.com/Vilamuzz/yota-backend/app/foster_children"
	"github.com/Vilamuzz/yota-backend/app/social_program"
)

type Repository interface {
	Create(ctx context.Context, record *FinanceRecord) error
	FindAll(ctx context.Context, options map[string]interface{}) ([]FinanceRecord, error)
	Summary(ctx context.Context, isAdmin bool) (FinanceRecordSummary, error)
	MonthlyTrend(ctx context.Context, params MonthlyTrendQueryParams) (FinanceMonthlyTrendResponse, error)
	Delete(ctx context.Context, id string) error
}

type repo struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repo{Conn: conn}
}

func (r *repo) Create(ctx context.Context, record *FinanceRecord) error {
	return r.Conn.WithContext(ctx).Create(record).Error
}

func (r *repo) FindAll(ctx context.Context, options map[string]interface{}) ([]FinanceRecord, error) {
	var records []FinanceRecord

	limit := options["limit"].(int)
	if limit <= 0 {
		limit = 10
	}

	usingPrevCursor := options["prev_cursor"] != ""

	var order string
	if usingPrevCursor {
		order = "created_at ASC, id ASC"
	} else {
		order = "created_at DESC, id DESC"
	}

	query := r.Conn.WithContext(ctx).Where("deleted_at IS NULL").Order(order).Limit(limit + 1)

	if options["fund_id"] != "" {
		query = query.Where("fund_id = ?", options["fund_id"])
	}
	if options["source_type"] != "" {
		query = query.Where("source_type = ?", options["source_type"])
	}
	if options["next_cursor"] != "" {
		cursorData, err := pkg.DecodeCursor(options["next_cursor"].(string))
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}
	if usingPrevCursor {
		cursorData, err := pkg.DecodeCursor(options["prev_cursor"].(string))
		if err == nil {
			query = query.Where("(created_at, id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	err := query.Find(&records).Error
	return records, err
}

func (r *repo) Summary(ctx context.Context, isAdmin bool) (FinanceRecordSummary, error) {
	var results []struct {
		FundType   string
		SourceType string
		Total      float64
	}

	err := r.Conn.WithContext(ctx).
		Model(&FinanceRecord{}).
		Where("deleted_at IS NULL").
		Select("fund_type, source_type, sum(amount) as total").
		Group("fund_type, source_type").
		Scan(&results).Error

	var summary FinanceRecordSummary
	if err != nil {
		return summary, err
	}

	var countDonationProgram int64
	r.Conn.WithContext(ctx).Model(&donation_program.DonationProgram{}).
		Where("status NOT IN ?", []string{string(donation_program.StatusDraft), string(donation_program.StatusArchived)}).
		Count(&countDonationProgram)
	summary.TotalDonationProgram = int(countDonationProgram)

	var countSocialProgram int64
	r.Conn.WithContext(ctx).Model(&social_program.SocialProgram{}).
		Where("status NOT IN ?", []string{string(social_program.StatusPending), string(social_program.StatusRejected)}).
		Count(&countSocialProgram)
	summary.TotalSocialProgram = int(countSocialProgram)

	var countFosterChildren int64
	r.Conn.WithContext(ctx).Model(&foster_children.FosterChildren{}).Count(&countFosterChildren)
	summary.TotalFosterChildren = int(countFosterChildren)

	for _, res := range results {
		if res.SourceType == SourceTypeExpense {
			switch res.FundType {
			case FundTypeDonation:
				summary.TotalDonationProgramExpense = res.Total
			case FundTypeSocialProgram:
				summary.TotalSocialProgramExpense = res.Total
			case FundTypeFosterChildren:
				summary.TotalFosterChildrenExpense = res.Total
			}
		} else if isAdmin && res.SourceType == SourceTypeTransaction {
			switch res.FundType {
			case FundTypeDonation:
				summary.TotalDonationProgramIncome = res.Total
			case FundTypeSocialProgram:
				summary.TotalSocialProgramIncome = res.Total
			case FundTypeFosterChildren:
				summary.TotalFosterChildrenIncome = res.Total
			}
		}
	}
	return summary, nil
}

func (r *repo) Delete(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&FinanceRecord{}).Error
}

func (r *repo) MonthlyTrend(ctx context.Context, params MonthlyTrendQueryParams) (FinanceMonthlyTrendResponse, error) {
	var response FinanceMonthlyTrendResponse
	response.Module = params.Module

	year := params.Year
	if year <= 0 {
		year = time.Now().Year()
	}

	type dbMonthlyTrend struct {
		MonthNum int     `gorm:"column:month_num"`
		Income   float64 `gorm:"column:income"`
		Expense  float64 `gorm:"column:expense"`
	}

	var dbResults []dbMonthlyTrend

	query := r.Conn.WithContext(ctx).
		Model(&FinanceRecord{}).
		Where("deleted_at IS NULL").
		Select("CAST(EXTRACT(MONTH FROM transaction_date) AS INTEGER) as month_num, "+
			"SUM(CASE WHEN source_type = ? THEN amount ELSE 0 END) as income, "+
			"SUM(CASE WHEN source_type = ? THEN amount ELSE 0 END) as expense",
			SourceTypeTransaction, SourceTypeExpense).
		Where("EXTRACT(YEAR FROM transaction_date) = ?", year)

	if params.Module != "" {
		query = query.Where("fund_type = ?", params.Module)
	}

	err := query.
		Group("month_num").
		Order("month_num ASC").
		Scan(&dbResults).Error

	if err != nil {
		return response, err
	}

	dbMap := make(map[int]dbMonthlyTrend)
	for _, res := range dbResults {
		dbMap[res.MonthNum] = res
	}

	response.Items = make([]FinanceMonthlyTrendItem, 12)
	for i := 1; i <= 12; i++ {
		monthStr := fmt.Sprintf("%d-%02d", year, i)
		item := FinanceMonthlyTrendItem{
			Month:   monthStr,
			Income:  0,
			Expense: 0,
		}

		if dbRes, exists := dbMap[i]; exists {
			item.Income = dbRes.Income
			item.Expense = dbRes.Expense
		}

		response.Items[i-1] = item
	}

	return response, nil
}
