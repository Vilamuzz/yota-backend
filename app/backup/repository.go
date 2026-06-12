package backup

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type repository struct {
	Conn *gorm.DB
}

type Repository interface {
	FindAllBackups(ctx context.Context) ([]Backup, error)
	FindOneBackup(ctx context.Context, id string) (*Backup, error)
	CreateBackup(ctx context.Context, backup *Backup) error
	DeleteBackup(ctx context.Context, id string) error
	GetOldBackups(ctx context.Context, cutoffTime time.Time) ([]Backup, error)
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FindAllBackups(ctx context.Context) ([]Backup, error) {
	var backups []Backup
	if err := r.Conn.WithContext(ctx).Where("deleted_at IS NULL").Order("created_at DESC").Find(&backups).Error; err != nil {
		return nil, err
	}
	return backups, nil
}

func (r *repository) FindOneBackup(ctx context.Context, id string) (*Backup, error) {
	var backup Backup
	if err := r.Conn.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&backup).Error; err != nil {
		return nil, err
	}
	return &backup, nil
}

func (r *repository) CreateBackup(ctx context.Context, backup *Backup) error {
	return r.Conn.WithContext(ctx).Create(backup).Error
}

func (r *repository) DeleteBackup(ctx context.Context, id string) error {
	now := time.Now()
	return r.Conn.WithContext(ctx).Model(&Backup{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

func (r *repository) GetOldBackups(ctx context.Context, cutoffTime time.Time) ([]Backup, error) {
	var backups []Backup
	if err := r.Conn.WithContext(ctx).Where("created_at < ? AND deleted_at IS NULL", cutoffTime).Find(&backups).Error; err != nil {
		return nil, err
	}
	return backups, nil
}
