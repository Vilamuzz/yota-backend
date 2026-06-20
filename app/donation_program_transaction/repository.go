package donation_program_transaction

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllDonationProgramTransactions(ctx context.Context, options map[string]interface{}) ([]DonationProgramTransaction, error)
	FindAllDonationProgramTransactionsForExport(ctx context.Context, donationProgramID string, params DonationProgramTransactionQueryParams) ([]DonationProgramTransaction, error)
	FindOneDonationProgramTransaction(ctx context.Context, options map[string]interface{}) (*DonationProgramTransaction, error)
	CreateDonationProgramTransaction(ctx context.Context, tx *DonationProgramTransaction) error
	UpdateDonationProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error
	CancelDonationProgramTransaction(ctx context.Context, orderID string) error
	GetMonthlyIncomeByProgram(ctx context.Context, donationProgramID string, year int) (*TransactionMonthlyIncomeRecord, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

var allowedDonationProgramTransactionSortColumns = map[string]string{
	"gross_amount": "gross_amount",
	"grossamount":  "gross_amount",
	"created_at":   "created_at",
	"createdat":    "created_at",
}

func (r *repository) FindAllDonationProgramTransactions(ctx context.Context, options map[string]interface{}) ([]DonationProgramTransaction, error) {
	var transactions []DonationProgramTransaction
	query := r.Conn.WithContext(ctx).Preload("DonationProgram")

	query = r.applyFilters(query, options)

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("(created_at, id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	if _, isPrev := options["prev_cursor"]; isPrev {
		query = query.Order("created_at ASC, id ASC")
	} else {
		orderClause := "created_at DESC, id DESC"
		if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
			parts := strings.Fields(strings.ToLower(sortBy.(string)))
			if len(parts) >= 1 {
				if col, valid := allowedDonationProgramTransactionSortColumns[parts[0]]; valid {
					dir := "ASC"
					if len(parts) == 2 && parts[1] == "desc" {
						dir = "DESC"
					}
					orderClause = fmt.Sprintf("%s %s, id DESC", col, dir)
				}
			}
		}
		query = query.Order(orderClause)
	}

	limit := 10
	if l, ok := options["limit"]; ok && l.(int) > 0 {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *repository) applyFilters(query *gorm.DB, options map[string]interface{}) *gorm.DB {
	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("transaction_status = ?", status.(string))
	}
	if donationID, ok := options["donation_program_id"]; ok && donationID.(string) != "" {
		query = query.Where("donation_program_id = ?", donationID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
	}
	if search, ok := options["search"]; ok && search.(string) != "" {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("donor_name ILIKE ? OR donor_email ILIKE ? OR order_id ILIKE ?", searchPattern, searchPattern, searchPattern)
	}
	if startDate, ok := options["start_date"]; ok && startDate.(string) != "" {
		query = query.Where("created_at >= ?", startDate.(string))
	}
	if endDate, ok := options["end_date"]; ok && endDate.(string) != "" {
		query = query.Where("created_at <= ?", endDate.(string)+" 23:59:59")
	}
	return query
}

func (r *repository) FindOneDonationProgramTransaction(ctx context.Context, options map[string]interface{}) (*DonationProgramTransaction, error) {
	var tx DonationProgramTransaction
	query := r.Conn.WithContext(ctx).Preload("DonationProgram")
	if id, ok := options["id"]; ok && id.(string) != "" {
		err := query.Where("id = ?", id.(string)).First(&tx).Error
		return &tx, err
	}
	if orderID, ok := options["order_id"]; ok && orderID.(string) != "" {
		err := query.Where("order_id = ?", orderID.(string)).First(&tx).Error
		return &tx, err
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		err := query.Where("account_id = ?", accountID.(string)).First(&tx).Error
		return &tx, err
	}
	if donationProgramID, ok := options["donation_program_id"]; ok && donationProgramID.(string) != "" {
		err := query.Where("donation_program_id = ?", donationProgramID.(string)).First(&tx).Error
		return &tx, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateDonationProgramTransaction(ctx context.Context, tx *DonationProgramTransaction) error {
	return r.Conn.WithContext(ctx).Create(tx).Error
}

func (r *repository) UpdateDonationProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&DonationProgramTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) CancelDonationProgramTransaction(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&DonationProgramTransaction{}).Where("id = ?", id).Update("transaction_status", "cancel").Error; err != nil {
			return err
		}

		if err := tx.Table("finance_records").Where("source_id = ? AND source_type = ?", id, "transaction").Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repository) GetMonthlyIncomeByProgram(ctx context.Context, donationProgramID string, year int) (*TransactionMonthlyIncomeRecord, error) {
	type dbMonthlyIncome struct {
		MonthNum int     `gorm:"column:month_num"`
		Income   float64 `gorm:"column:income"`
	}

	var dbResults []dbMonthlyIncome

	err := r.Conn.WithContext(ctx).
		Model(&DonationProgramTransaction{}).
		Select("CAST(EXTRACT(MONTH FROM paid_at) AS INTEGER) as month_num, SUM(gross_amount) as income").
		Where("donation_program_id = ?", donationProgramID).
		Where("EXTRACT(YEAR FROM paid_at) = ?", year).
		Where("transaction_status = ? OR (transaction_status = ? AND fraud_status != ?)", "settlement", "capture", "challenge").
		Group("month_num").
		Order("month_num ASC").
		Scan(&dbResults).Error

	if err != nil {
		return nil, err
	}

	dbMap := make(map[int]float64)
	for _, res := range dbResults {
		dbMap[res.MonthNum] = res.Income
	}

	record := &TransactionMonthlyIncomeRecord{
		DonationProgramID: donationProgramID,
		Items:             make([]TransactionMonthlyIncomeItem, 12),
	}

	for i := 1; i <= 12; i++ {
		monthStr := fmt.Sprintf("%d-%02d", year, i)
		income := 0.0
		if val, exists := dbMap[i]; exists {
			income = val
		}
		record.Items[i-1] = TransactionMonthlyIncomeItem{
			Month:  monthStr,
			Income: income,
		}
	}

	return record, nil
}

func (r *repository) FindAllDonationProgramTransactionsForExport(ctx context.Context, donationProgramID string, params DonationProgramTransactionQueryParams) ([]DonationProgramTransaction, error) {
	var transactions []DonationProgramTransaction
	query := r.Conn.WithContext(ctx).Order("created_at ASC, id ASC")
	if donationProgramID != "" {
		query = query.Where("donation_program_id = ?", donationProgramID)
	}
	if params.Status != "" {
		query = query.Where("transaction_status = ?", params.Status)
	}
	if params.StartDate != "" {
		query = query.Where("created_at >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("created_at <= ?", params.EndDate+" 23:59:59")
	}
	err := query.Find(&transactions).Error
	return transactions, err
}
