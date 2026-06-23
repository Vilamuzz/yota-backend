package foster_children_expense

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllFosterChildrenExpenses(ctx context.Context, options map[string]interface{}) ([]FosterChildrenExpense, error)
	FindAllFosterChildrenExpensesForExport(ctx context.Context, fosterChildrenSlug string, params FosterChildrenExpenseExportParams) ([]FosterChildrenExpense, error)
	FindOneFosterChildrenExpense(ctx context.Context, options map[string]interface{}) (*FosterChildrenExpense, error)
	GetTotalExpenseByFosterChildrenID(ctx context.Context, fosterChildrenID string) (float64, error)
	CreateFosterChildrenExpense(ctx context.Context, fosterChildrenExpense *FosterChildrenExpense) error
	DeleteFosterChildrenExpense(ctx context.Context, fosterChildrenExpenseID string) error
}

type repo struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repo{Conn: conn}
}

var allowedFosterChildrenExpenseSortColumns = map[string]string{
	"title":        "foster_children_expenses.title",
	"amount":       "foster_children_expenses.amount",
	"expense_date": "foster_children_expenses.expense_date",
	"created_at":   "foster_children_expenses.created_at",
}

func (r *repo) FindAllFosterChildrenExpenses(ctx context.Context, options map[string]interface{}) ([]FosterChildrenExpense, error) {
	var expenses []FosterChildrenExpense

	query := r.Conn.WithContext(ctx).Where("foster_children_expenses.deleted_at IS NULL")

	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		query = query.Where("foster_children_expenses.foster_children_id = ?", fosterChildrenID.(string))
	}

	if fosterChildrenSlug, ok := options["foster_children_slug"]; ok && fosterChildrenSlug.(string) != "" {
		query = query.Joins("JOIN foster_childrens ON foster_childrens.id = foster_children_expenses.foster_children_id").
			Where("foster_childrens.slug = ?", fosterChildrenSlug.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("foster_children_expenses.title ILIKE ?", "%"+search.(string)+"%")
	}

	if startDate, ok := options["start_date"]; ok && startDate.(string) != "" {
		query = query.Where("foster_children_expenses.expense_date >= ?", startDate.(string))
	}

	if endDate, ok := options["end_date"]; ok && endDate.(string) != "" {
		query = query.Where("foster_children_expenses.expense_date <= ?", endDate.(string))
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("(foster_children_expenses.created_at, foster_children_expenses.id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor.(string) != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("(foster_children_expenses.created_at, foster_children_expenses.id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		orderClause := "foster_children_expenses.created_at DESC, foster_children_expenses.id DESC"
		if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
			parts := strings.Fields(strings.ToLower(sortBy.(string)))
			if len(parts) >= 1 {
				if col, valid := allowedFosterChildrenExpenseSortColumns[parts[0]]; valid {
					dir := "ASC"
					if len(parts) == 2 && parts[1] == "desc" {
						dir = "DESC"
					}
					orderClause = fmt.Sprintf("%s %s, foster_children_expenses.id DESC", col, dir)
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
	err := query.Find(&expenses).Error
	return expenses, err
}

func (r *repo) FindAllFosterChildrenExpensesForExport(ctx context.Context, fosterChildrenSlug string, params FosterChildrenExpenseExportParams) ([]FosterChildrenExpense, error) {
	var expenses []FosterChildrenExpense
	query := r.Conn.WithContext(ctx).Order("foster_children_expenses.expense_date ASC, foster_children_expenses.created_at ASC").
		Where("foster_children_expenses.deleted_at IS NULL")
	if fosterChildrenSlug != "" {
		query = query.Joins("JOIN foster_childrens ON foster_childrens.id = foster_children_expenses.foster_children_id").
			Where("foster_childrens.slug = ?", fosterChildrenSlug)
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

func (r *repo) FindOneFosterChildrenExpense(ctx context.Context, options map[string]interface{}) (*FosterChildrenExpense, error) {
	var expense FosterChildrenExpense
	if id, ok := options["id"]; ok && id != "" {
		err := r.Conn.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&expense).Error
		return &expense, err
	}
	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("foster_children_id = ? AND deleted_at IS NULL", fosterChildrenID.(string)).First(&expense).Error
		return &expense, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repo) CreateFosterChildrenExpense(ctx context.Context, expense *FosterChildrenExpense) error {
	return r.Conn.WithContext(ctx).Create(expense).Error
}

func (r *repo) DeleteFosterChildrenExpense(ctx context.Context, fosterChildrenExpenseID string) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&FosterChildrenExpense{}).Where("id = ?", fosterChildrenExpenseID).Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		if err := tx.Table("finance_records").Where("source_id = ? AND source_type = ?", fosterChildrenExpenseID, "expense").Update("deleted_at", time.Now()).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repo) GetTotalExpenseByFosterChildrenID(ctx context.Context, fosterChildrenID string) (float64, error) {
	var total float64
	err := r.Conn.WithContext(ctx).
		Table("foster_children_expenses").
		Where("foster_children_id = ? AND deleted_at IS NULL", fosterChildrenID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
