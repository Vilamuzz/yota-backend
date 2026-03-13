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
	FindAll(ctx context.Context, options QueryParams) ([]DonationExpense, error)
	Update(ctx context.Context, expense *DonationExpense) error
	Delete(ctx context.Context, id string) error
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, expense *DonationExpense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *repo) FindByID(ctx context.Context, id string) (*DonationExpense, error) {
	var expense DonationExpense
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&expense).Error
	return &expense, err
}

func (r *repo) FindAll(ctx context.Context, options QueryParams) ([]DonationExpense, error) {
	var expenses []DonationExpense

	usingPrevCursor := options.PrevCursor != ""

	order := "created_at DESC, id DESC"
	if usingPrevCursor {
		order = "created_at ASC, id ASC"
	}

	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}

	query := r.db.WithContext(ctx).Order(order).Limit(limit + 1)

	if options.DonationID != "" {
		query = query.Where("donation_id = ?", options.DonationID)
	}
	if options.NextCursor != "" {
		cursorData, err := pkg.DecodeCursor(options.NextCursor)
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}
	if usingPrevCursor {
		cursorData, err := pkg.DecodeCursor(options.PrevCursor)
		if err == nil {
			query = query.Where("(created_at, id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	err := query.Find(&expenses).Error
	return expenses, err
}

func (r *repo) Update(ctx context.Context, expense *DonationExpense) error {
	expense.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(expense).Error
}

func (r *repo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&DonationExpense{}).Error
}
