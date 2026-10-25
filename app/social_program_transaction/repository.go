package social_program_transaction

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/app/finance_record"
	"github.com/Vilamuzz/yota-backend/app/social_program_invoice"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	FindAllSocialProgramTransactions(ctx context.Context, options map[string]interface{}) ([]SocialProgramTransaction, error)
	FindOneSocialProgramTransaction(ctx context.Context, options map[string]interface{}) (*SocialProgramTransaction, error)
	CreateSocialProgramTransaction(ctx context.Context, transaction *SocialProgramTransaction) error
	CreateOfflinePayment(ctx context.Context, transaction *SocialProgramTransaction, record *finance_record.FinanceRecord) error
	UpdateSocialProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error
	ApplyNotification(ctx context.Context, orderID, newStatus string, updates map[string]interface{}, record *finance_record.FinanceRecord) (bool, error)
	WithTransaction(ctx context.Context, fn func(repo Repository) error) error
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
	if socialProgramInvoiceID, ok := options["social_program_invoice_id"]; ok && socialProgramInvoiceID.(string) != "" {
		query = query.Where("social_program_invoice_id = ?", socialProgramInvoiceID.(string))
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

func (r *repository) CreateOfflinePayment(ctx context.Context, transaction *SocialProgramTransaction, record *finance_record.FinanceRecord) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(transaction).Error; err != nil {
			return err
		}

		var invoice social_program_invoice.SocialProgramInvoice
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", transaction.SocialProgramInvoiceID).First(&invoice).Error; err != nil {
			return err
		}
		if err := tx.Model(&invoice).Updates(map[string]interface{}{
			"status":     social_program_invoice.InvoiceStatusPaid,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&social_program_subscription.SocialProgramSubscription{}).
			Where("id = ?", invoice.SubscriptionID).
			UpdateColumns(map[string]interface{}{
				"total_paid_periods": gorm.Expr("total_paid_periods + 1"),
				"updated_at":         time.Now(),
			}).Error; err != nil {
			return err
		}

		if record != nil {
			if err := finance_record.CreateIdempotent(tx, record); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) ApplyNotification(ctx context.Context, orderID, newStatus string, updates map[string]interface{}, record *finance_record.FinanceRecord) (bool, error) {
	applied := false
	err := r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transaction SocialProgramTransaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("order_id = ?", orderID).First(&transaction).Error; err != nil {
			return err
		}

		if transaction.TransactionStatus == newStatus {
			return nil
		}

		if err := tx.Model(&transaction).Updates(updates).Error; err != nil {
			return err
		}

		if record != nil {
			var invoice social_program_invoice.SocialProgramInvoice
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", transaction.SocialProgramInvoiceID).First(&invoice).Error; err != nil {
				return err
			}
			if invoice.Status != social_program_invoice.InvoiceStatusPaid {
				if err := tx.Model(&invoice).Updates(map[string]interface{}{
					"status":     social_program_invoice.InvoiceStatusPaid,
					"updated_at": time.Now(),
				}).Error; err != nil {
					return err
				}
				if err := tx.Model(&social_program_subscription.SocialProgramSubscription{}).
					Where("id = ?", invoice.SubscriptionID).
					UpdateColumns(map[string]interface{}{
						"total_paid_periods": gorm.Expr("total_paid_periods + 1"),
						"updated_at":         time.Now(),
					}).Error; err != nil {
					return err
				}
			}
			if err := finance_record.CreateIdempotent(tx, record); err != nil {
				return err
			}
		}

		applied = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return applied, nil
}

func (r *repository) UpdateSocialProgramTransaction(ctx context.Context, orderID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgramTransaction{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) WithTransaction(ctx context.Context, fn func(repo Repository) error) error {
	return r.Conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repository{Conn: tx})
	})
}
