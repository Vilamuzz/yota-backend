package social_program

import (
	"context"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllSocialPrograms(ctx context.Context, options map[string]interface{}) ([]SocialProgram, error)
	FindOneSocialProgram(ctx context.Context, options map[string]interface{}) (*SocialProgram, error)
	CreateSocialProgram(ctx context.Context, socialProgram *SocialProgram) error
	UpdateSocialProgram(ctx context.Context, socialProgramID string, updates map[string]interface{}) error
	DeleteSocialProgram(ctx context.Context, socialProgramID string) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAllSocialPrograms(ctx context.Context, options map[string]interface{}) ([]SocialProgram, error) {
	var socialPrograms []SocialProgram
	query := r.Conn.WithContext(ctx)

	if status, ok := options["status"]; ok && status.(string) != "" {
		query = query.Where("status = ?", status.(string))
	}

	if search, ok := options["search"]; ok && search.(string) != "" {
		query = query.Where("title ILIKE ?", "%"+search.(string)+"%")
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
	if err := query.Find(&socialPrograms).Error; err != nil {
		return nil, err
	}
	return socialPrograms, nil
}

func (r *repository) FindOneSocialProgram(ctx context.Context, options map[string]interface{}) (*SocialProgram, error) {
	var socialProgram SocialProgram
	if id, ok := options["id"]; ok && id.(string) != "" {
		err := r.Conn.WithContext(ctx).Where("id = ?", id.(string)).First(&socialProgram).Error
		return &socialProgram, err
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *repository) CreateSocialProgram(ctx context.Context, socialProgram *SocialProgram) error {
	return r.Conn.WithContext(ctx).Create(socialProgram).Error
}

func (r *repository) UpdateSocialProgram(ctx context.Context, socialProgramID string, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&SocialProgram{}).
		Where("id = ?", socialProgramID).
		Updates(updates).Error
}

func (r *repository) DeleteSocialProgram(ctx context.Context, socialProgramID string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", socialProgramID).Update("deleted_at", time.Now()).Error
}
