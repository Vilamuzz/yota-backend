package ambulance_service_request

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, ambulanceServiceRequest AmbulanceServiceRequest) error
	FindByID(ctx context.Context, id string) (AmbulanceServiceRequest, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceServiceRequest, error)
	Count(ctx context.Context, options map[string]interface{}) (int64, error)
	Update(ctx context.Context, id string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

var allowedAmbulanceServiceRequestSortColumns = map[string]string{
	"applicant_name": "applicant_name",
	"created_at":     "created_at",
}

func (r *repository) Create(ctx context.Context, ambulanceServiceRequest AmbulanceServiceRequest) error {
	return r.Conn.Create(&ambulanceServiceRequest).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (AmbulanceServiceRequest, error) {
	var ambulanceServiceRequest AmbulanceServiceRequest
	if err := r.Conn.WithContext(ctx).
		Preload("Ambulance.Driver.UserProfile").
		First(&ambulanceServiceRequest, "id = ?", id).Error; err != nil {
		return AmbulanceServiceRequest{}, err
	}
	return ambulanceServiceRequest, nil
}

func (r *repository) Count(ctx context.Context, options map[string]interface{}) (int64, error) {
	var count int64
	query := r.Conn.WithContext(ctx).Model(&AmbulanceServiceRequest{})

	if accountID, ok := options["account_id"]; ok && accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	if ambulanceID, ok := options["ambulance_id"]; ok && ambulanceID != "" {
		query = query.Where("ambulance_id = ?", ambulanceID)
	}

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if serviceCategory, ok := options["service_category"]; ok && serviceCategory != "" {
		query = query.Where("service_category = ?", serviceCategory)
	}

	if search, ok := options["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		if onlyName, ok := options["search_only_name"]; ok && onlyName.(bool) {
			query = query.Where("applicant_name ILIKE ?", searchTerm)
		} else {
			query = query.Where("applicant_name ILIKE ? OR applicant_phone ILIKE ?", searchTerm, searchTerm)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]AmbulanceServiceRequest, error) {
	var ambulanceServiceRequests []AmbulanceServiceRequest
	query := r.Conn.WithContext(ctx)

	if accountID, ok := options["account_id"]; ok && accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}

	if ambulanceID, ok := options["ambulance_id"]; ok && ambulanceID != "" {
		query = query.Where("ambulance_id = ?", ambulanceID)
	}

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if serviceCategory, ok := options["service_category"]; ok && serviceCategory != "" {
		query = query.Where("service_category = ?", serviceCategory)
	}

	if search, ok := options["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		if onlyName, ok := options["search_only_name"]; ok && onlyName.(bool) {
			query = query.Where("applicant_name ILIKE ?", searchTerm)
		} else {
			query = query.Where("applicant_name ILIKE ? OR applicant_phone ILIKE ?", searchTerm, searchTerm)
		}
	}

	if nextCursor, ok := options["next_cursor"]; ok && nextCursor != "" {
		cursorData, err := pkg.DecodeCursor(nextCursor.(string))
		if err == nil {
			query = query.Where("created_at < ? OR (created_at = ? AND id < ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID)
		}
	} else if prevCursor, ok := options["prev_cursor"]; ok && prevCursor != "" {
		cursorData, err := pkg.DecodeCursor(prevCursor.(string))
		if err == nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)",
				cursorData.CreatedAt, cursorData.CreatedAt, cursorData.ID).
				Order("created_at ASC, id ASC")
		}
	}

	if _, usingPrevCursor := options["prev_cursor"]; !usingPrevCursor {
		orderClause := "created_at DESC, id DESC"
		if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
			parts := strings.Fields(strings.ToLower(sortBy.(string)))
			if len(parts) >= 1 {
				if col, valid := allowedAmbulanceServiceRequestSortColumns[parts[0]]; valid {
					dir := "ASC"
					if len(parts) == 2 && parts[1] == "desc" {
						dir = "DESC"
					}
					orderClause = fmt.Sprintf("%s %s", col, dir)
				}
			}
		}
		query = query.Order(orderClause)
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	if page, ok := options["page"]; ok {
		p := page.(int)
		if p <= 0 {
			p = 1
		}
		offset := (p - 1) * limit
		query = query.Limit(limit).Offset(offset)
	} else {
		query = query.Limit(limit + 1)
	}

	if err := query.Find(&ambulanceServiceRequests).Error; err != nil {
		return nil, err
	}
	return ambulanceServiceRequests, nil
}

func (r *repository) Update(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.Model(&AmbulanceServiceRequest{}).Where("id = ?", id).Updates(updateData).Error
}

