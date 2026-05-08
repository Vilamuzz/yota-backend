package social_program_invoice

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
	"github.com/Vilamuzz/yota-backend/app/social_program_subscription"
)

type Repository interface {
	FindAllSocialProgramInvoices(ctx context.Context, options map[string]interface{}) ([]SocialProgramInvoice, error)
	FindOneSocialProgramInvoice(ctx context.Context, options map[string]interface{}) (*SocialProgramInvoice, error)
	CreateSocialProgramInvoice(ctx context.Context, socialProgramInvoice *SocialProgramInvoice) error
	UpdateSocialProgramInvoice(ctx context.Context, socialProgramInvoiceID string, updates map[string]interface{}) error
	DeleteSocialProgramInvoice(ctx context.Context, socialProgramInvoiceID string) error
	FindActiveSubscriptionsForBillingDay(ctx context.Context, billingDay int) ([]social_program_subscription.SocialProgramSubscription, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllSocialProgramInvoices(ctx context.Context, options map[string]interface{}) ([]SocialProgramInvoice, error) {
	var invoices []SocialProgramInvoice
	query := r.Conn.WithContext(ctx)

	if subscriptionID, ok := options["subscription_id"]; ok && subscriptionID.(string) != "" {
		query = query.Where("subscription_id = ?", subscriptionID.(string))
	}

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("status = ?", status.(string))
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
	if err := query.Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *repository) FindOneSocialProgramInvoice(ctx context.Context, options map[string]interface{}) (*SocialProgramInvoice, error) {
	var invoice SocialProgramInvoice
	query := r.Conn.WithContext(ctx)

	if id, ok := options["id"]; ok && id.(string) != "" {
		query = query.Where("id = ?", id.(string))
	}
	
	err := query.First(&invoice).Error
	return &invoice, err
}

func (r *repository) CreateSocialProgramInvoice(ctx context.Context, socialProgramInvoice *SocialProgramInvoice) error {
	return r.Conn.WithContext(ctx).Create(socialProgramInvoice).Error
}

func (r *repository) UpdateSocialProgramInvoice(ctx context.Context, socialProgramInvoiceID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgramInvoice{}).
		Where("id = ?", socialProgramInvoiceID).
		Updates(updates).Error
}

func (r *repository) DeleteSocialProgramInvoice(ctx context.Context, socialProgramInvoiceID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", socialProgramInvoiceID).Delete(&SocialProgramInvoice{}).Error
}

func (r *repository) FindActiveSubscriptionsForBillingDay(ctx context.Context, billingDay int) ([]social_program_subscription.SocialProgramSubscription, error) {
	var subscriptions []social_program_subscription.SocialProgramSubscription
	err := r.Conn.WithContext(ctx).
		Joins("JOIN social_programs ON social_programs.id = social_program_subscriptions.social_program_id").
		Where("social_programs.billing_day = ?", billingDay).
		Where("social_program_subscriptions.status != ?", "tidak_aktif").
		Find(&subscriptions).Error
	return subscriptions, err
}