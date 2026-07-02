package foster_children

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	FindAllFosterChildren(ctx context.Context, options map[string]interface{}) ([]FosterChildren, error)
	CountFosterChildren(ctx context.Context, options map[string]interface{}) (int64, error)
	FindOneFosterChildren(ctx context.Context, options map[string]interface{}) (*FosterChildren, error)
	CreateFosterChildren(ctx context.Context, fosterChildren *FosterChildren) error
	UpdateFosterChildren(ctx context.Context, fosterChildrenID string, updateData map[string]interface{}) error
	DeleteFosterChildren(ctx context.Context, fosterChildrenID string) error
	DeleteAchievementsByFosterChildrenID(ctx context.Context, fosterChildrenID string) error
	DeleteAchievementByID(ctx context.Context, id string) error
	UpdateAchievement(ctx context.Context, id string, updateData map[string]interface{}) error
	CreateAchievements(ctx context.Context, achievements []Achivement) error
	CreateFosterChildrenFromCandidate(ctx context.Context, candidateID, name, nik, profilePicture, familyCard, sktm, birthPlace, schoolName, address string, gender string, category string, educationLevel int, birthDate time.Time) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

// allowedFosterChildrenSortColumns whitelists sortable columns to prevent SQL injection.
var allowedFosterChildrenSortColumns = map[string]string{
	"name":       "name",
	"birth_date": "birth_date",
	"category":   "category",
	"gender":     "gender",
	"created_at": "created_at",
}

func (r *repository) FindAllFosterChildren(ctx context.Context, options map[string]interface{}) ([]FosterChildren, error) {
	var fosterChildren []FosterChildren
	query := r.Conn.WithContext(ctx)
	totalExpenseSubquery := r.Conn.Table("foster_children_expenses").
		Select("COALESCE(SUM(amount), 0)").
		Where("foster_children_id = foster_childrens.id")
	query = query.Select("foster_childrens.*, (?) as total_expense", totalExpenseSubquery)

	if isAdmin, ok := options["is_admin"].(bool); ok && isAdmin {
		collectedFundSubquery := r.Conn.Table("foster_children_transactions").
			Select("COALESCE(SUM(gross_amount), 0)").
			Where("foster_children_id = foster_childrens.id AND transaction_status = 'settlement'")

		query = query.Select("foster_childrens.*, (?) as collected_fund", collectedFundSubquery)
	}

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

	// Build ORDER BY from "sort_by" option, e.g. "name asc" or "education_level desc".
	orderClause := "created_at DESC"
	if sortBy, ok := options["sort_by"]; ok && sortBy.(string) != "" {
		parts := strings.Fields(strings.ToLower(sortBy.(string)))
		if len(parts) >= 1 {
			if col, valid := allowedFosterChildrenSortColumns[parts[0]]; valid {
				dir := "ASC"
				if len(parts) == 2 && parts[1] == "desc" {
					dir = "DESC"
				}
				orderClause = fmt.Sprintf("%s %s", col, dir)
			}
		}
	}
	query = query.Order(orderClause)

	limit := 10
	if l, ok := options["limit"]; ok && l.(int) > 0 {
		limit = l.(int)
	}
	offset := 0
	if page, ok := options["page"]; ok && page.(int) > 1 {
		offset = (page.(int) - 1) * limit
	}

	query = query.Limit(limit).Offset(offset)
	if err := query.Find(&fosterChildren).Error; err != nil {
		return nil, err
	}
	return fosterChildren, nil
}

func (r *repository) CountFosterChildren(ctx context.Context, options map[string]interface{}) (int64, error) {
	var total int64
	query := r.Conn.WithContext(ctx).Model(&FosterChildren{})
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
	err := query.Count(&total).Error
	return total, err
}

func (r *repository) FindOneFosterChildren(ctx context.Context, options map[string]interface{}) (*FosterChildren, error) {
	var fosterChildren FosterChildren

	collectedFundSubquery := r.Conn.Table("foster_children_transactions").
		Select("COALESCE(SUM(gross_amount), 0)").
		Where("foster_children_id = foster_childrens.id AND transaction_status = 'settlement'")

	totalExpenseSubquery := r.Conn.Table("foster_children_expenses").
		Select("COALESCE(SUM(amount), 0)").
		Where("foster_children_id = foster_childrens.id")

	query := r.Conn.WithContext(ctx).
		Select("foster_childrens.*, (?) as collected_fund, (?) as total_expense", collectedFundSubquery, totalExpenseSubquery).
		Preload("Achivements")

	if id, ok := options["id"]; ok && id != "" {
		query = query.Where("id = ?", id)
	} else if slug, ok := options["slug"]; ok && slug != "" {
		query = query.Where("slug = ?", slug)
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

func (r *repository) DeleteAchievementByID(ctx context.Context, id string) error {
	return r.Conn.WithContext(ctx).Where("id = ?", id).Delete(&Achivement{}).Error
}

func (r *repository) UpdateAchievement(ctx context.Context, id string, updateData map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&Achivement{}).Where("id = ?", id).Updates(updateData).Error
}

func (r *repository) CreateAchievements(ctx context.Context, achievements []Achivement) error {
	return r.Conn.WithContext(ctx).Create(&achievements).Error
}

func (r *repository) CreateFosterChildrenFromCandidate(ctx context.Context, candidateID, name, nik, profilePicture, familyCard, sktm, birthPlace, schoolName, address string, gender string, category string, educationLevel int, birthDate time.Time) error {
	now := time.Now()
	fc := &FosterChildren{
		ID:             uuid.New(),
		Slug:           fmt.Sprintf("%s-%s", pkg.Slugify(name), uuid.New().String()[:5]),
		Name:           name,
		Nik:            nik,
		ProfilePicture: profilePicture,
		Gender:         Gender(gender),
		IsGraduated:    false,
		Category:       Category(category),
		BirthDate:      birthDate,
		BirthPlace:     birthPlace,
		SchoolName:     schoolName,
		EducationLevel: educationLevel,
		Address:        address,
		FamilyCard:     familyCard,
		SKTM:           sktm,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	return r.Conn.WithContext(ctx).Create(fc).Error
}
