package social_program_expense

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialProgramExpenses(ctx context.Context, options map[string]interface{}) ([]SocialProgramExpense, error)
	FindAllSocialProgramExpensesForExport(ctx context.Context, socialProgramID string, params SocialProgramExpenseExportParams) ([]SocialProgramExpense, error)
	FindOneSocialProgramExpense(ctx context.Context, options map[string]interface{}) (*SocialProgramExpense, error)
	CreateSocialProgramExpense(ctx context.Context, socialProgramExpense *SocialProgramExpense) error
	DeleteSocialProgramExpense(ctx context.Context, socialProgramExpenseID string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

var allowedSocialProgramExpenseSortColumns = map[string]string{
	"title":        "title",
	"amount":       "amount",
	"expense_date": "expense_date",
	"created_at":   "created_at",
}

func (r *repository) FindAllSocialProgramExpenses(ctx context.Context, options map[string]interface{}) ([]SocialProgramExpense, error) {
	var expenses []SocialProgramExpense
	query := r.Conn.WithContext(ctx).Where("deleted_at IS NULL")

	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
	}

	if startDate, ok := options["start_date"]; ok && startDate.(string) != "" {
		query = query.Where("expense_date >= ?", startDate.(string))
	}

	if endDate, ok := options["end_date"]; ok && endDate.(string) != "" {
		query = query.Where("expense_date <= ?", endDate.(string))
	}

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

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		orderClause := "created_at DESC, id DESC"
		if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
			parts := strings.Fields(strings.ToLower(sortBy.(string)))
			if len(parts) >= 1 {
				if col, valid := allowedSocialProgramExpenseSortColumns[parts[0]]; valid {
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
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&expenses).Error; err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *repository) FindAllSocialProgramExpensesForExport(ctx context.Context, socialProgramID string, params SocialProgramExpenseExportParams) ([]SocialProgramExpense, error) {
	var expenses []SocialProgramExpense
	query := r.Conn.WithContext(ctx).Order("expense_date ASC, created_at ASC").Where("deleted_at IS NULL")
	if socialProgramID != "" {
		query = query.Where("social_program_id = ?", socialProgramID)
	}
	if params.StartDate != "" {
		query = query.Where("expense_date >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("expense_date <= ?", params.EndDate)
	}
	err := query.Find(&expenses).Error
	return expenses, err
}

func (r *repository) FindOneSocialProgramExpense(ctx context.Context, options map[string]interface{}) (*SocialProgramExpense, error) {
	var expense SocialProgramExpense
	if id, ok := options["id"]; ok && id.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id.(string)).First(&expense).Error
		return &expense, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateSocialProgramExpense(ctx context.Context, socialProgramExpense *SocialProgramExpense) error {
	return r.Conn.WithContext(ctx).Create(socialProgramExpense).Error
}

func (r *repository) DeleteSocialProgramExpense(ctx context.Context, socialProgramExpenseID string) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&SocialProgramExpense{}).Where("id = ?", socialProgramExpenseID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		if err := tx.Table("finance_records").Where("source_id = ? AND source_type = ?", socialProgramExpenseID, "expense").Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})
}
