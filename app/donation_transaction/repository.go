package donation_transaction

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tx *DonationTransaction) error
	FindByID(ctx context.Context, id string) (*DonationTransaction, error)
	FindByOrderID(ctx context.Context, orderID string) (*DonationTransaction, error)
	UpdateStatus(ctx context.Context, orderID string, updates map[string]interface{}) error
	FindAll(ctx context.Context, params QueryParams) ([]DonationTransaction, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, tx *DonationTransaction) error {
	return r.Conn.WithContext(ctx).Create(tx).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*DonationTransaction, error) {
	var tx DonationTransaction
	if err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repository) FindByOrderID(ctx context.Context, orderID string) (*DonationTransaction, error) {
	var tx DonationTransaction
	if err := r.Conn.WithContext(ctx).Where("order_id = ?", orderID).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repository) UpdateStatus(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&DonationTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) FindAll(ctx context.Context, params QueryParams) ([]DonationTransaction, error) {
	var transactions []DonationTransaction

	usingPrevCursor := params.PrevCursor != ""

	order := "created_at DESC, id DESC"
	if usingPrevCursor {
		order = "created_at ASC, id ASC"
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 10
	}

	query := r.Conn.WithContext(ctx).Order(order).Limit(limit + 1)

	if params.Status != "" {
		query = query.Where("transaction_status = ?", params.Status)
	}
	if params.DonationID != "" {
		query = query.Where("donation_id = ?", params.DonationID)
	}
	if params.NextCursor != "" {
		cursorData, err := pkg.DecodeCursor(params.NextCursor)
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}
	if usingPrevCursor {
		cursorData, err := pkg.DecodeCursor(params.PrevCursor)
		if err == nil {
			query = query.Where("(created_at, id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
