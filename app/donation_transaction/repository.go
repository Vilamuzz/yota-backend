package donation_transaction

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tx *DonationTransaction) error
	FindByID(ctx context.Context, id string) (*DonationTransaction, error)
	FindByOrderID(ctx context.Context, orderID string) (*DonationTransaction, error)
	UpdateStatus(ctx context.Context, orderID string, updates map[string]interface{}) error
	FindAll(ctx context.Context, options map[string]interface{}) ([]DonationTransaction, error)
	FindByDonationID(ctx context.Context, donationID string) ([]DonationTransaction, error)
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

func (r *repository) FindByDonationID(ctx context.Context, donationID string) ([]DonationTransaction, error) {
	var transactions []DonationTransaction
	if err := r.Conn.WithContext(ctx).Where("donation_id = ?", donationID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
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

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("transaction_status = ?", status)
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
