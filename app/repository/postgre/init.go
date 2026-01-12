package postgre_repository

import (
	"github.com/Vilamuzz/yota-backend/domain"
	"gorm.io/gorm"
)

type postgreDbRepo struct {
	Conn *gorm.DB
}

func NewPostgreDBRepo(conn *gorm.DB) domain.PostgreDBRepository {
	return &postgreDbRepo{Conn: conn}
}
