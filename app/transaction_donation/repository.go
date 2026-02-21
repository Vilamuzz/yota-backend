package transaction_donation

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tx *TransactionDonation) error
	FindByID(ctx context.Context, id string) (*TransactionDonation, error)
	FindByOrderID(ctx context.Context, orderID string) (*TransactionDonation, error)
	UpdateStatus(ctx context.Context, orderID string, status string, transactionID string, paidAt *time.Time) error
	FindAll(ctx context.Context, options map[string]interface{}) ([]TransactionDonation, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, tx *TransactionDonation) error {
	return r.Conn.WithContext(ctx).Create(tx).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*TransactionDonation, error) {
	var tx TransactionDonation
	if err := r.Conn.WithContext(ctx).Where("id = ?", id).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repository) FindByOrderID(ctx context.Context, orderID string) (*TransactionDonation, error) {
	var tx TransactionDonation
	if err := r.Conn.WithContext(ctx).Where("order_id = ?", orderID).First(&tx).Error; err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repository) UpdateStatus(ctx context.Context, orderID string, status string, transactionID string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"payment_status": status,
		"updated_at":     time.Now(),
	}
	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}
	return r.Conn.WithContext(ctx).Model(&TransactionDonation{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]TransactionDonation, error) {
	var transactions []TransactionDonation
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("payment_status = ?", status)
	}
	if donationID, ok := options["donation_id"]; ok && donationID != "" {
		query = query.Where("donation_id = ?", donationID)
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	query = query.Limit(limit + 1)
	query = query.Order("created_at DESC")

	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}
