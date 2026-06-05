package social_program_subscription

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error)
	CountSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) (int64, error)
	FindOneSocialProgramSubscription(ctx context.Context, options map[string]interface{}) (*SocialProgramSubscription, error)
	CreateSocialProgramSubscription(ctx context.Context, subscription *SocialProgramSubscription) error
	UpdateSocialProgramSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) error
	DeleteSocialProgramSubscription(ctx context.Context, subscriptionID string) error
	FindSubscriptionsDueForBilling(ctx context.Context, billingDay int) ([]SocialProgramSubscription, error)
	FindAllSubscribers(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error)
	CountSubscribers(ctx context.Context, options map[string]interface{}) (int64, error)
	GetSubscriberStats(ctx context.Context, accountIDs []string) (map[string]SubscriberStats, error)
	GetTotalDonationBySubscriptionIDs(ctx context.Context, subscriptionIDs []string) (map[string]float64, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error) {
	var subscriptions []SocialProgramSubscription
	query := r.Conn.WithContext(ctx).Table("social_program_subscriptions").
		Select(`social_program_subscriptions.*,
			(SELECT COALESCE(SUM(spt.gross_amount), 0)
			 FROM social_program_transactions spt
			 JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id
			 WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') as total_donation,
			(SELECT COUNT(*)
			 FROM social_program_transactions spt
			 JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id
			 WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') as total_paid_periods`).
		Preload("Account.UserProfile").Preload("SocialProgram")

	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
	}
	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("status = ?", status.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		searchQuery := "%" + search.(string) + "%"
		if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
			query = query.Joins("JOIN social_programs ON social_programs.id = social_program_subscriptions.social_program_id").
				Where("social_programs.title ILIKE ?", searchQuery)
		} else {
			query = query.Joins("JOIN accounts ON accounts.id = social_program_subscriptions.account_id").
				Joins("JOIN user_profiles ON user_profiles.account_id = accounts.id").
				Where("user_profiles.username ILIKE ? OR accounts.email ILIKE ?", searchQuery, searchQuery)
		}
	}

	orderClause := "social_program_subscriptions.created_at DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		switch sortBy.(string) {
		case "total_donation desc":
			orderClause = "(SELECT COALESCE(SUM(spt.gross_amount), 0) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') DESC"
		case "total_donation asc":
			orderClause = "(SELECT COALESCE(SUM(spt.gross_amount), 0) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') ASC"
		case "total_paid_periods desc":
			orderClause = "(SELECT COUNT(*) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') DESC"
		case "total_paid_periods asc":
			orderClause = "(SELECT COUNT(*) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') ASC"
		case "created_at asc":
			orderClause = "social_program_subscriptions.created_at ASC"
		}
	}
	query = query.Order(orderClause)

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *repository) CountSocialProgramSubscriptions(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&SocialProgramSubscription{})

	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_id = ?", socialProgramID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("account_id = ?", accountID.(string))
	}
	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("status = ?", status.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		searchQuery := "%" + search.(string) + "%"
		if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
			query = query.Joins("JOIN social_programs ON social_programs.id = social_program_subscriptions.social_program_id").
				Where("social_programs.title ILIKE ?", searchQuery)
		} else {
			query = query.Joins("JOIN accounts ON accounts.id = social_program_subscriptions.account_id").
				Joins("JOIN user_profiles ON user_profiles.account_id = accounts.id").
				Where("user_profiles.username ILIKE ? OR accounts.email ILIKE ?", searchQuery, searchQuery)
		}
	}

	err := query.Count(&total).Error
	return total, err
}

func (r *repository) FindOneSocialProgramSubscription(ctx context.Context, options map[string]interface{}) (*SocialProgramSubscription, error) {
	var subscription SocialProgramSubscription
	query := r.Conn.WithContext(ctx).Table("social_program_subscriptions").
		Select(`social_program_subscriptions.*,
			(SELECT COALESCE(SUM(spt.gross_amount), 0)
			 FROM social_program_transactions spt
			 JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id
			 WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') as total_donation`).
		Preload("Account.UserProfile").Preload("SocialProgram")

	if id, ok := options["id"]; ok && id.(string) != "" {
		query = query.Where("social_program_subscriptions.id = ?", id.(string))
	}
	if socialProgramID, ok := options["social_program_id"]; ok && socialProgramID.(string) != "" {
		query = query.Where("social_program_subscriptions.social_program_id = ?", socialProgramID.(string))
	}
	if accountID, ok := options["account_id"]; ok && accountID.(string) != "" {
		query = query.Where("social_program_subscriptions.account_id = ?", accountID.(string))
	}
	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("social_program_subscriptions.status = ?", status.(string))
	}

	if err := query.First(&subscription).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
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

func (r *repository) FindSubscriptionsDueForBilling(ctx context.Context, billingDay int) ([]SocialProgramSubscription, error) {
	var subscriptions []SocialProgramSubscription
	err := r.Conn.WithContext(ctx).
		Joins("JOIN social_programs ON social_programs.id = social_program_subscriptions.social_program_id").
		Where("social_program_subscriptions.status = ?", StatusActive).
		Where("social_programs.status = ?", "active").
		Where("social_programs.billing_day = ?", billingDay).
		Preload("SocialProgram").
		Find(&subscriptions).Error
	return subscriptions, err
}

func (r *repository) FindAllSubscribers(ctx context.Context, options map[string]interface{}) ([]SocialProgramSubscription, error) {
	var subscriptions []SocialProgramSubscription

	// Subquery to get one representative subscription per account
	subQuery := r.Conn.Model(&SocialProgramSubscription{}).
		Select("MAX(id)").
		Group("account_id")

	query := r.Conn.WithContext(ctx).Model(&SocialProgramSubscription{}).
		Where("social_program_subscriptions.id IN (?)", subQuery).
		Preload("Account.UserProfile")

	if search, ok := options["search"]; ok && search.(string) != "" {
		searchQuery := "%" + search.(string) + "%"
		query = query.Joins("JOIN accounts ON accounts.id = social_program_subscriptions.account_id").
			Joins("JOIN user_profiles ON user_profiles.account_id = accounts.id").
			Where("user_profiles.username ILIKE ? OR accounts.email ILIKE ?", searchQuery, searchQuery)
	}

	orderClause := "social_program_subscriptions.created_at DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		// Just in case we need sort by for subscribers list
		switch sortBy.(string) {
		case "total_donation desc":
			orderClause = "(SELECT COALESCE(SUM(spt.gross_amount), 0) FROM social_program_transactions spt WHERE spt.account_id = social_program_subscriptions.account_id AND spt.transaction_status = 'settlement') DESC"
		case "total_donation asc":
			orderClause = "(SELECT COALESCE(SUM(spt.gross_amount), 0) FROM social_program_transactions spt WHERE spt.account_id = social_program_subscriptions.account_id AND spt.transaction_status = 'settlement') ASC"
		case "total_subscription desc":
			orderClause = "(SELECT COUNT(*) FROM social_program_subscriptions sps WHERE sps.account_id = social_program_subscriptions.account_id) DESC"
		case "total_subscription asc":
			orderClause = "(SELECT COUNT(*) FROM social_program_subscriptions sps WHERE sps.account_id = social_program_subscriptions.account_id) ASC"
		case "total_paid_periods desc":
			orderClause = "(SELECT COUNT(*) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') DESC"
		case "total_paid_periods asc":
			orderClause = "(SELECT COUNT(*) FROM social_program_transactions spt JOIN social_program_invoices spi ON spi.id = spt.social_program_invoice_id WHERE spi.subscription_id = social_program_subscriptions.id AND spt.transaction_status = 'settlement') ASC"
		case "created_at asc":
			orderClause = "social_program_subscriptions.created_at ASC"
		}
	}
	query = query.Order(orderClause)

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}
	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	return subscriptions, nil
}

func (r *repository) CountSubscribers(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	subQuery := r.Conn.Model(&SocialProgramSubscription{}).
		Select("MAX(id)").
		Group("account_id")

	query := r.Conn.WithContext(ctx).Model(&SocialProgramSubscription{}).
		Where("social_program_subscriptions.id IN (?)", subQuery)

	if search, ok := options["search"]; ok && search.(string) != "" {
		searchQuery := "%" + search.(string) + "%"
		query = query.Joins("JOIN accounts ON accounts.id = social_program_subscriptions.account_id").
			Joins("JOIN user_profiles ON user_profiles.account_id = accounts.id").
			Where("user_profiles.username ILIKE ? OR accounts.email ILIKE ?", searchQuery, searchQuery)
	}

	err := query.Count(&total).Error
	return total, err
}

type SubscriberStats struct {
	AccountID         string
	TotalSubscription int
	TotalDonation     float64
}

func (r *repository) GetSubscriberStats(ctx context.Context, accountIDs []string) (map[string]SubscriberStats, error) {
	if len(accountIDs) == 0 {
		return make(map[string]SubscriberStats), nil
	}

	stats := make(map[string]SubscriberStats)

	type subCount struct {
		AccountID string
		Count     int
	}
	var subCounts []subCount
	err := r.Conn.WithContext(ctx).
		Table("social_program_subscriptions").
		Select("account_id, COUNT(id) as count").
		Where("account_id IN ?", accountIDs).
		Group("account_id").
		Find(&subCounts).Error
	if err != nil {
		return nil, err
	}

	type donSum struct {
		AccountID string
		Total     float64
	}
	var donSums []donSum
	err = r.Conn.WithContext(ctx).
		Table("social_program_transactions").
		Select("account_id, SUM(gross_amount) as total").
		Where("account_id IN ? AND transaction_status = ?", accountIDs, "settlement").
		Group("account_id").
		Find(&donSums).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	for _, sc := range subCounts {
		stats[sc.AccountID] = SubscriberStats{
			AccountID:         sc.AccountID,
			TotalSubscription: sc.Count,
		}
	}
	for _, ds := range donSums {
		s := stats[ds.AccountID]
		s.TotalDonation = float64(ds.Total)
		stats[ds.AccountID] = s
	}

	return stats, nil
}

func (r *repository) GetTotalDonationBySubscriptionIDs(ctx context.Context, subscriptionIDs []string) (map[string]float64, error) {
	if len(subscriptionIDs) == 0 {
		return make(map[string]float64), nil
	}

	donations := make(map[string]float64)

	type donSum struct {
		SubscriptionID string
		Total          float64
	}
	var donSums []donSum
	err := r.Conn.WithContext(ctx).
		Table("social_program_transactions").
		Select("social_program_invoices.subscription_id as subscription_id, SUM(social_program_transactions.gross_amount) as total").
		Joins("JOIN social_program_invoices ON social_program_invoices.id = social_program_transactions.social_program_invoice_id").
		Where("social_program_invoices.subscription_id IN ? AND social_program_transactions.transaction_status = ?", subscriptionIDs, "settlement").
		Group("social_program_invoices.subscription_id").
		Find(&donSums).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	for _, ds := range donSums {
		donations[ds.SubscriptionID] = float64(ds.Total)
	}

	return donations, nil
}
