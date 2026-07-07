package payment

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]PaymentMethods, error)
	FindByID(ctx context.Context, id int) (*PaymentMethods, error)
	Update(ctx context.Context, id int, updates map[string]interface{}) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}

func (r *repository) FindAll(ctx context.Context) ([]PaymentMethods, error) {
	var list []PaymentMethods
	err := r.Conn.WithContext(ctx).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *repository) FindByID(ctx context.Context, id int) (*PaymentMethods, error) {
	var pm PaymentMethods
	err := r.Conn.WithContext(ctx).First(&pm, id).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *repository) Update(ctx context.Context, id int, updates map[string]interface{}) error {
	return r.Conn.WithContext(ctx).Model(&PaymentMethods{}).Where("id = ?", id).Updates(updates).Error
}
