package social_program_transaction

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialProgramTransactions(ctx context.Context, options map[string]interface{}) ([]SocialProgramTransaction, error)
	FindOneSocialProgramTransaction(ctx context.Context, options map[string]interface{}) (*SocialProgramTransaction, error)
	CreateSocialProgramTransaction(ctx context.Context, transaction *SocialProgramTransaction) error
	UpdateSocialProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllSocialProgramTransactions(ctx context.Context, options map[string]interface{}) ([]SocialProgramTransaction, error) {
	var transactions []SocialProgramTransaction
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("transaction_status = ?", status.(string))
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

func (r *repository) FindOneSocialProgramTransaction(ctx context.Context, options map[string]interface{}) (*SocialProgramTransaction, error) {
	var transaction SocialProgramTransaction
	query := r.Conn.WithContext(ctx)

	if id, ok := options["id"]; ok && id.(string) != "" {
		query = query.Where("id = ?", id.(string))
	}
	if orderID, ok := options["order_id"]; ok && orderID.(string) != "" {
		query = query.Where("order_id = ?", orderID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
	}

	err := query.First(&transaction).Error
	return &transaction, err
}

func (r *repository) CreateSocialProgramTransaction(ctx context.Context, transaction *SocialProgramTransaction) error {
	return r.Conn.WithContext(ctx).Create(transaction).Error
}

func (r *repository) UpdateSocialProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgramTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}