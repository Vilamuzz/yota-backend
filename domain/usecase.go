package domain

import (
	"context"

	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type SuperadminAppUsecase interface {
	LoginSuperadmin(ctx context.Context, req request.UserLoginRequest) pkg.Response
}

type AdminAppUsecase interface {
	LoginAdmin(ctx context.Context, req request.UserLoginRequest) pkg.Response
}

type UserAppUsecase interface {
	RegisterUser(ctx context.Context, req request.UserRegisterRequest) pkg.Response
	LoginUser(ctx context.Context, req request.UserLoginRequest) pkg.Response
}
