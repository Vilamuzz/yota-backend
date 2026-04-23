package foster_children

import (
	"context"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllFosterChildren(ctx context.Context, options map[string]interface{}) ([]FosterChildren, error)
	FindOneFosterChildren(ctx context.Context, options map[string]interface{}) (*FosterChildren, error)
	CreateFosterChildren(ctx context.Context, fosterChildren *FosterChildren) error
	UpdateFosterChildren(ctx context.Context, fosterChildrenID string, updateData map[string]interface{}) error
	DeleteFosterChildren(ctx context.Context, fosterChildrenID string) error
	DeleteAchievementsByFosterChildrenID(ctx context.Context, fosterChildrenID string) error
	CreateAchievements(ctx context.Context, achievements []Achivement) error

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

func (r *repository) FindAllFosterChildren(ctx context.Context, options map[string]interface{}) ([]FosterChildren, error) {
	var fosterChildren []FosterChildren
	query := r.Conn.WithContext(ctx).
		Preload("Achivements")

	if search, ok := options["search"]; ok && search != "" {
		query = query.Where("name ILIKE ?", "%"+search.(string)+"%")
	}
	if category, ok := options["category"]; ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if gender, ok := options["gender"]; ok && gender != "" {
		query = query.Where("gender = ?", gender)
	}
	if isGraduated, ok := options["is_graduated"]; ok {
		query = query.Where("is_graduated = ?", isGraduated)
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
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
	if err := query.Find(&fosterChildren).Error; err != nil {
		return nil, err
	}
	return fosterChildren, nil
}

func (r *repository) FindOneFosterChildren(ctx context.Context, options map[string]interface{}) (*FosterChildren, error) {
	var fosterChildren FosterChildren
	query := r.Conn.WithContext(ctx).
		Preload("Achivements")

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	}

	if err := query.First(&fosterChildren).Error; err != nil {
		return nil, err
	}
	return &fosterChildren, nil
}

func (r *repository) CreateFosterChildren(ctx context.Context, fosterChildren *FosterChildren) error {
	return r.Conn.WithContext(ctx).Create(fosterChildren).Error
}

func (r *repository) UpdateFosterChildren(ctx context.Context, fosterChildrenID string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&FosterChildren{}).Where("id = ?", fosterChildrenID).Updates(updateData).Error
}

func (r *repository) DeleteFosterChildren(ctx context.Context, fosterChildrenID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", fosterChildrenID).Delete(&FosterChildren{}).Error
}

func (r *repository) DeleteAchievementsByFosterChildrenID(ctx context.Context, fosterChildrenID string) error {
	return r.Conn.WithContext(ctx).Where("foster_children_id = ?", fosterChildrenID).Delete(&Achivement{}).Error
}

func (r *repository) CreateAchievements(ctx context.Context, achievements []Achivement) error {
	return r.Conn.WithContext(ctx).Create(&achievements).Error
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
		query = query.Order("created_at DESC, id DESC")
	}

	limit := 10
	if l, ok := options["limit"]; ok {
		limit = l.(int)
	}

	query = query.Limit(limit + 1)
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
