package news

import "gorm.io/gorm"

type Repository interface {
	FetchAllNews() ([]News, error)
	Create(news *News) error
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{
		Conn: conn,
	}
}

func (r *repository) FetchAllNews() ([]News, error) {
	var newsList []News
	if err := r.Conn.Find(&newsList).Error; err != nil {
		return nil, err
	}
	return newsList, nil
}

func (r *repository) Create(news *News) error {
	return r.Conn.Create(news).Error
}
