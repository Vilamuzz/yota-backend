package social_program_expense

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialProgramExpenses(ctx context.Context, options map[string]interface{}) ([]SocialProgramExpense, error)
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

func (r *repository) FindAllSocialProgramExpenses(ctx context.Context, options map[string]interface{}) ([]SocialProgramExpense, error) {
	var expenses []SocialProgramExpense
	query := r.Conn.WithContext(ctx)

	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
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
	if err := query.Find(&expenses).Error; err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *repository) FindOneSocialProgramExpense(ctx context.Context, options map[string]interface{}) (*SocialProgramExpense, error) {
	var expense SocialProgramExpense
	if id, ok := options["id"]; ok && id.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("id = ?", id.(string)).First(&expense).Error
		return &expense, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateSocialProgramExpense(ctx context.Context, socialProgramExpense *SocialProgramExpense) error {
	return r.Conn.WithContext(ctx).Create(socialProgramExpense).Error
}

func (r *repository) DeleteSocialProgramExpense(ctx context.Context, socialProgramExpenseID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", socialProgramExpenseID).Delete(&SocialProgramExpense{}).Error
}
