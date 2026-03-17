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
	FindAll(ctx context.Context, options map[string]interface{}) ([]DonationTransaction, error)
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

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]DonationTransaction, error) {
	var transactions []DonationTransaction
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("transaction_status = ?", status.(string))
	}
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
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
