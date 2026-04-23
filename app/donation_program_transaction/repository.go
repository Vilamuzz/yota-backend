package donation_program_transaction

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllDonationProgramTransactions(ctx context.Context, options map[string]interface{}) ([]DonationProgramTransaction, error)
	FindOneDonationProgramTransaction(ctx context.Context, options map[string]interface{}) (*DonationProgramTransaction, error)
	CreateDonationProgramTransaction(ctx context.Context, tx *DonationProgramTransaction) error
	UpdateDonationProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllDonationProgramTransactions(ctx context.Context, options map[string]interface{}) ([]DonationProgramTransaction, error) {
	var transactions []DonationProgramTransaction
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("transaction_status = ?", status.(string))
	}
	if donationID, ok := options["donation_program_id"]; ok && donationID.(string) != "" {
		query = query.Where("donation_program_id = ?", donationID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
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

func (r *repository) FindOneDonationProgramTransaction(ctx context.Context, options map[string]interface{}) (*DonationProgramTransaction, error) {
	var tx DonationProgramTransaction
	if id, ok := options["id"]; ok && id.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("id = ?", id.(string)).First(&tx).Error
		return &tx, err
	}
	if orderID, ok := options["order_id"]; ok && orderID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("order_id = ?", orderID.(string)).First(&tx).Error
		return &tx, err
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("account_id = ?", accountID.(string)).First(&tx).Error
		return &tx, err
	}
	if donationProgramID, ok := options["donation_program_id"]; ok && donationProgramID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("donation_program_id = ?", donationProgramID.(string)).First(&tx).Error
		return &tx, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateDonationProgramTransaction(ctx context.Context, tx *DonationProgramTransaction) error {
	return r.Conn.WithContext(ctx).Create(tx).Error
}

func (r *repository) UpdateDonationProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&DonationProgramTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}
