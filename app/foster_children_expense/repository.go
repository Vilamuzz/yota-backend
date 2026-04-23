package foster_children_expense

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllFosterChildrenExpenses(ctx context.Context, options map[string]interface{}) ([]FosterChildrenExpense, error)
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

func (r *repo) FindAllFosterChildrenExpenses(ctx context.Context, options map[string]interface{}) ([]FosterChildrenExpense, error) {
	var expenses []FosterChildrenExpense

	query := r.Conn.WithContext(ctx)

	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		query = query.Where("foster_children_id = ?", fosterChildrenID.(string))
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
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	err := query.Find(&expenses).Error
	return expenses, err
}

func (r *repo) FindOneFosterChildrenExpense(ctx context.Context, options map[string]interface{}) (*FosterChildrenExpense, error) {
	var expense FosterChildrenExpense
	if id, ok := options["id"]; ok && id != "" {
		err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&expense).Error
		return &expense, err
	}
	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("foster_children_id = ?", fosterChildrenID.(string)).First(&expense).Error
		return &expense, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repo) CreateFosterChildrenExpense(ctx context.Context, expense *FosterChildrenExpense) error {
	return r.Conn.WithContext(ctx).Create(expense).Error
}

func (r *repo) DeleteFosterChildrenExpense(ctx context.Context, fosterChildrenExpenseID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", fosterChildrenExpenseID).Delete(&FosterChildrenExpense{}).Error
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
