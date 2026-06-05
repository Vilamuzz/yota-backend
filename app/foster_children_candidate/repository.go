package foster_children_candidate

import (
	"context"
	"fmt"
	"strings"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	CountFosterChildrenCandidates(ctx context.Context, options map[string]interface{}) (int64, error)
	FindAllFosterChildrenCandidates(ctx context.Context, options map[string]interface{}) ([]FosterChildrenCandidate, error)
	FindOneFosterChildrenCandidate(ctx context.Context, options map[string]interface{}) (*FosterChildrenCandidate, error)
	CreateFosterChildrenCandidate(ctx context.Context, candidate *FosterChildrenCandidate) error
	UpdateFosterChildrenCandidate(ctx context.Context, id string, updateData map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

var allowedFosterChildrenCandidateSortColumns = map[string]string{
	"name":       "name",
	"created_at": "created_at",
}

func (r *repository) CountFosterChildrenCandidates(ctx context.Context, options map[string]interface{}) (int64, error) {
	var count int64
	query := r.Conn.WithContext(ctx).Model(&FosterChildrenCandidate{})

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if accountID, ok := options["account_id"]; ok && accountID != "" {
		query = query.Where("submitted_by = ?", accountID)
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if gender, ok := options["gender"]; ok && gender != "" {
		query = query.Where("gender = ?", gender)
	}
	if search, ok := options["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		query = query.Where("name ILIKE ? OR submitter_name ILIKE ?", searchTerm, searchTerm)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) FindAllFosterChildrenCandidates(ctx context.Context, options map[string]interface{}) ([]FosterChildrenCandidate, error) {
	var candidates []FosterChildrenCandidate
	query := r.Conn.WithContext(ctx).Preload("Account").Preload("Account.UserProfile")

	if status, ok := options["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if accountID, ok := options["account_id"]; ok && accountID != "" {
		query = query.Where("submitted_by = ?", accountID)
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if gender, ok := options["gender"]; ok && gender != "" {
		query = query.Where("gender = ?", gender)
	}
	if search, ok := options["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		query = query.Where("name ILIKE ? OR submitter_name ILIKE ?", searchTerm, searchTerm)
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
				if col, valid := allowedFosterChildrenCandidateSortColumns[parts[0]]; valid {
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
		// Offset pagination
		p := page.(int)
		if p <= 0 {
			p = 1
		}
		offset := (p - 1) * limit
		query = query.Limit(limit).Offset(offset)
	} else {
		// Cursor pagination
		query = query.Limit(limit + 1)
	}

	if err := query.Find(&candidates).Error; err != nil {
		return nil, err
	}
	return candidates, nil
}

func (r *repository) FindOneFosterChildrenCandidate(ctx context.Context, options map[string]interface{}) (*FosterChildrenCandidate, error) {
	var candidate FosterChildrenCandidate
	query := r.Conn.WithContext(ctx).Preload("Account").Preload("Account.UserProfile")

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&candidate).Error; err != nil {
		return nil, err
	}
	return &candidate, nil
}

func (r *repository) CreateFosterChildrenCandidate(ctx context.Context, candidate *FosterChildrenCandidate) error {
	return r.Conn.WithContext(ctx).Create(candidate).Error
}

func (r *repository) UpdateFosterChildrenCandidate(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&FosterChildrenCandidate{}).
		Where("id = ?", id).
		Updates(updateData).Error
}
