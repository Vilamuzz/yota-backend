package foster_child_transaction

import "gorm.io/gorm"

type Repository interface {
}

type repository struct {
	Conn *gorm.DB
}

func NewRepository(conn *gorm.DB) Repository {
	return &repository{Conn: conn}
}
