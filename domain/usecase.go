package domain

import (
	"context"

	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
)

type SuperadminAppUsecase interface {
	LoginSuperadmin(ctx context.Context, req request.LoginRequest) pkg.Response
}

type AdminAppUsecase interface {
	LoginAdmin(ctx context.Context, req request.LoginRequest) pkg.Response
}

type UserAppUsecase interface {
	RegisterUser(ctx context.Context, req request.RegisterRequest) pkg.Response
	LoginUser(ctx context.Context, req request.LoginRequest) pkg.Response
	ForgetPassword(ctx context.Context, req request.ForgetPasswordRequest) pkg.Response
	ResetPassword(ctx context.Context, req request.ResetPasswordRequest) pkg.Response
}
