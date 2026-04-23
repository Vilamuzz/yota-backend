package social_program_subscription

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error)
	FindOneSocialProgramSubscription(ctx context.Context, options map[string]interface{}) (*SocialProgramSubscription, error)
	CreateSocialProgramSubscription(ctx context.Context, subscription *SocialProgramSubscription) error
	UpdateSocialProgramSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) error
	DeleteSocialProgramSubscription(ctx context.Context, subscriptionID string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error) {
	var subscriptions []SocialProgramSubscription
	query := r.Conn.WithContext(ctx)

	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
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
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *repository) FindOneSocialProgramSubscription(ctx context.Context, options map[string]interface{}) (*SocialProgramSubscription, error) {
	var subscription SocialProgramSubscription
	query := r.Conn.WithContext(ctx)

	if id, ok := options["id"]; ok && id.(string) != "" {
		query = query.Where("id = ?", id.(string))
	}
	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
	}

	err := query.First(&subscription).Error
	return &subscription, err
}

func (r *repository) CreateSocialProgramSubscription(ctx context.Context, subscription *SocialProgramSubscription) error {
	return r.Conn.WithContext(ctx).Create(subscription).Error
}

func (r *repository) UpdateSocialProgramSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgramSubscription{}).
		Where("id = ?", subscriptionID).
		Updates(updates).Error
}

func (r *repository) DeleteSocialProgramSubscription(ctx context.Context, subscriptionID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", subscriptionID).Delete(&SocialProgramSubscription{}).Error
}
