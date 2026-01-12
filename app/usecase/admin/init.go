package admin_usecase

import (
	"time"

	"github.com/Vilamuzz/yota-backend/domain"
)

type adminAppUsecase struct {
	postgreDbRepo  domain.PostgreDBRepository
	contextTimeout time.Duration
}

type RepoInjection struct {
	PostgreDBRepo domain.PostgreDBRepository
}

func NewAdminAppUsecase(repoInjection *RepoInjection, timeout time.Duration) domain.AdminAppUsecase {
	return &adminAppUsecase{
		postgreDbRepo:  repoInjection.PostgreDBRepo,
		contextTimeout: timeout,
	}
}
