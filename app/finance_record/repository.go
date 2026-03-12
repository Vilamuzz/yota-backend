package finance_record

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, record *FinanceRecord) error
	FindAll(ctx context.Context, options map[string]interface{}) ([]FinanceRecord, error)
}

type repo struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repo{Conn: conn}
}

func (r *repo) Create(ctx context.Context, record *FinanceRecord) error {
	return r.Conn.WithContext(ctx).Create(record).Error
}

func (r *repo) FindAll(ctx context.Context, options map[string]interface{}) ([]FinanceRecord, error) {
	var records []FinanceRecord

	limit := options["limit"].(int)
	if limit <= 0 {
		limit = 10
	}

	usingPrevCursor := options["prev_cursor"] != ""

	var order string
	if usingPrevCursor {
		order = "created_at ASC, id ASC"
	} else {
		order = "created_at DESC, id DESC"
	}

	query := r.Conn.WithContext(ctx).Order(order).Limit(limit + 1)

	if options["fund_id"] != "" {
		query = query.Where("fund_id = ?", options["fund_id"])
	}
	if options["source_type"] != "" {
		query = query.Where("source_type = ?", options["source_type"])
	}
	if options["next_cursor"] != "" {
		cursorData, err := pkg.DecodeCursor(options["next_cursor"].(string))
		if err == nil {
			query = query.Where("(created_at, id) < (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}
	if usingPrevCursor {
		cursorData, err := pkg.DecodeCursor(options["prev_cursor"].(string))
		if err == nil {
			query = query.Where("(created_at, id) > (?, ?)", cursorData.CreatedAt, cursorData.ID)
		}
	}

	err := query.Find(&records).Error
	return records, err
}
