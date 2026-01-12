package superadmin_usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Vilamuzz/yota-backend/config"
	postgre_model "github.com/Vilamuzz/yota-backend/domain/model/postgre"
	"github.com/Vilamuzz/yota-backend/domain/request"
	"github.com/Vilamuzz/yota-backend/pkg"
	jwt_pkg "github.com/Vilamuzz/yota-backend/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (u *superadminAppUsecase) LoginSuperadmin(ctx context.Context, req request.UserLoginRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	errValidation := make(map[string]string)
	if req.Email == "" {
		errValidation["email"] = "Email is required"
	}
	if req.Password == "" {
		errValidation["password"] = "Password is required"
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusUnprocessableEntity, "Validation Error", errValidation, nil)
	}

	superadmin, err := u.postgreDbRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email, "role": postgre_model.RoleSuperadmin})
	if superadmin == nil {
		return pkg.NewResponse(http.StatusUnauthorized, "User not found", nil, nil)
	}
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(superadmin.Password), []byte(req.Password))
	if err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Wrong password", nil, nil)
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(config.GetJWTTTL()) * time.Minute)
	token, err := jwt_pkg.GenerateJWTTokenUser(jwt_pkg.UserJWTClaims{
		UserID: superadmin.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "superadmin",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	})
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}

	data := map[string]interface{}{"token": token, "expired_at": expiredAt, "user": superadmin}
	return pkg.NewResponse(http.StatusOK, "Login successful", nil, data)
}
