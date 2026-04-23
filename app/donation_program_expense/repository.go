package donation_program_expense

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllDonationProgramExpenses(ctx context.Context, options map[string]interface{}) ([]DonationProgramExpense, error)
	FindOneDonationProgramExpense(ctx context.Context, options map[string]interface{}) (*DonationProgramExpense, error)
	GetTotalExpenseByDonationProgramID(ctx context.Context, donationProgramID string) (float64, error)
	CreateDonationProgramExpense(ctx context.Context, donationProgramExpense *DonationProgramExpense) error
	DeleteDonationProgramExpense(ctx context.Context, donationProgramExpenseID string) error
}

type repo struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repo{Conn: conn}
}

func (r *repo) FindAllDonationProgramExpenses(ctx context.Context, options map[string]interface{}) ([]DonationProgramExpense, error) {
	var expenses []DonationProgramExpense

	query := r.Conn.WithContext(ctx)

	if donationProgramID, ok := options["donation_program_id"]; ok && donationProgramID.(string) != "" {
		query = query.Where("donation_program_id = ?", donationProgramID.(string))
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

func (r *repo) FindOneDonationProgramExpense(ctx context.Context, options map[string]interface{}) (*DonationProgramExpense, error) {
	var expense DonationProgramExpense
	if id, ok := options["id"]; ok && id != "" {
		err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&expense).Error
		return &expense, err
	}
	if donationProgramID, ok := options["donation_program_id"]; ok && donationProgramID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("donation_program_id = ?", donationProgramID.(string)).First(&expense).Error
		return &expense, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repo) CreateDonationProgramExpense(ctx context.Context, expense *DonationProgramExpense) error {
	return r.Conn.WithContext(ctx).Create(expense).Error
}

func (r *repo) DeleteDonationProgramExpense(ctx context.Context, donationProgramExpenseID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", donationProgramExpenseID).Delete(&DonationProgramExpense{}).Error
}

func (r *repo) GetTotalExpenseByDonationProgramID(ctx context.Context, donationProgramID string) (float64, error) {
	var total float64
	err := r.Conn.WithContext(ctx).
		Table("donation_program_expenses").
		Where("donation_program_id = ? AND deleted_at IS NULL", donationProgramID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
