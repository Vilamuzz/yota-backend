package foster_children_transaction

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllFosterChildrenTransactions(ctx context.Context, options map[string]interface{}) ([]FosterChildrenTransaction, error)
	FindOneFosterChildrenTransaction(ctx context.Context, options map[string]interface{}) (*FosterChildrenTransaction, error)
	CreateFosterChildrenTransaction(ctx context.Context, tx *FosterChildrenTransaction) error
	UpdateFosterChildrenTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllFosterChildrenTransactions(ctx context.Context, options map[string]interface{}) ([]FosterChildrenTransaction, error) {
	var transactions []FosterChildrenTransaction
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("transaction_status = ?", status.(string))
	}
	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		query = query.Where("foster_children_id = ?", fosterChildrenID.(string))
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

func (r *repository) FindOneFosterChildrenTransaction(ctx context.Context, options map[string]interface{}) (*FosterChildrenTransaction, error) {
	var tx FosterChildrenTransaction
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
	if fosterChildrenID, ok := options["foster_children_id"]; ok && fosterChildrenID.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("foster_children_id = ?", fosterChildrenID.(string)).First(&tx).Error
		return &tx, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateFosterChildrenTransaction(ctx context.Context, tx *FosterChildrenTransaction) error {
	return r.Conn.WithContext(ctx).Create(tx).Error
}

func (r *repository) UpdateFosterChildrenTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&FosterChildrenTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}
