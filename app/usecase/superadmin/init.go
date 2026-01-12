package superadmin_usecase

import (
	"time"

	"github.com/Vilamuzz/yota-backend/domain"
)

type superadminAppUsecase struct {
	postgreDbRepo  domain.PostgreDBRepository
	contextTimeout time.Duration
}

type RepoInjection struct {
	PostgreDBRepo domain.PostgreDBRepository
}

func NewSuperadminAppUsecase(repoInjection *RepoInjection, timeout time.Duration) domain.SuperadminAppUsecase {
	return &superadminAppUsecase{
		postgreDbRepo:  repoInjection.PostgreDBRepo,
		contextTimeout: timeout,
	}
}
