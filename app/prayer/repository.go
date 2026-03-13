package prayer

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, prayer *Prayer) error
	FindByID(ctx context.Context, id string) (*Prayer, error)
	FindAll(ctx context.Context, options map[string]interface{}) ([]Prayer, error)
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) Create(ctx context.Context, prayer *Prayer) error {
	return r.Conn.Create(prayer).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*Prayer, error) {
	var prayer Prayer
	if err := r.Conn.First(&prayer, id).Error; err != nil {
		return nil, err
	}
	return &prayer, nil
}

func (r *repository) FindAll(ctx context.Context, options map[string]interface{}) ([]Prayer, error) {
	var prayers []Prayer
	if err := r.Conn.Find(&prayers).Error; err != nil {
		return nil, err
	}
	return prayers, nil
}
