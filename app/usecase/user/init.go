package user_usecase

import (
	"time"

	"github.com/Vilamuzz/yota-backend/domain"
)

type userAppUsecase struct {
	postgreDbRepo  domain.PostgreDBRepository
	contextTimeout time.Duration
}

type RepoInjection struct {
	PostgreDBRepo domain.PostgreDBRepository
}

func NewUserAppUsecase(repoInjection *RepoInjection, timeout time.Duration) domain.UserAppUsecase {
	return &userAppUsecase{
		postgreDbRepo:  repoInjection.PostgreDBRepo,
		contextTimeout: timeout,
	}
}
