package donation_expense

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, expense *DonationExpense) error
	FindByID(ctx context.Context, id string) (*DonationExpense, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]DonationExpense, error)
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	GetTotalExpenseByDonationID(ctx context.Context, donationID string) (float64, error)
}

type repo struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repo{Conn: conn}
}

func (r *repo) Create(ctx context.Context, expense *DonationExpense) error {
	return r.Conn.WithContext(ctx).Create(expense).Error
}

func (r *repo) FindByID(ctx context.Context, id string) (*DonationExpense, error) {
	var expense DonationExpense
	err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&expense).Error
	return &expense, err
}

func (r *repo) FindAll(ctx context.Context, options map[string]interface{}) ([]DonationExpense, error) {
	var expenses []DonationExpense

	query := r.Conn.WithContext(ctx)

	if donationID, ok := options["donation_id"]; ok && donationID.(string) != "" {
		query = query.Where("donation_id = ?", donationID.(string))
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

func (r *repo) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	updateData["updated_at"] = time.Now()
	return r.Conn.WithContext(ctx).Model(&DonationExpense{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repo) Delete(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&DonationExpense{}).Error
}

func (r *repo) GetTotalExpenseByDonationID(ctx context.Context, donationID string) (float64, error) {
	var total float64
	err := r.Conn.WithContext(ctx).
		Table("donation_expenses").
		Where("donation_id = ? AND deleted_at IS NULL", donationID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}
