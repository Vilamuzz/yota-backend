package user_usecase

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
	"gorm.io/gorm"
)

func (u *userAppUsecase) RegisterUser(ctx context.Context, req request.UserRegisterRequest) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	errValidation := make(map[string]string)
	if req.Username == "" {
		errValidation["username"] = "Username is required"
	}
	if req.Email == "" {
		errValidation["email"] = "Email is required"
	}
	if req.Password == "" {
		errValidation["password"] = "Password is required"
	} else {
		if !pkg.IsValidLengthPassword(req.Password) {
			errValidation["password"] = "Password must be at least 8 characters long"
		}
		if !pkg.IsStrongPassword(req.Password) {
			errValidation["password"] = "Password must contain at least one uppercase letter, one lowercase letter, and one digit"
		}
	}
	if len(errValidation) > 0 {
		return pkg.NewResponse(http.StatusUnprocessableEntity, "Validation Error", errValidation, nil)
	}

	user, err := u.postgreDbRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if user != nil {
		return pkg.NewResponse(http.StatusConflict, "User with that email already exists", nil, nil)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, "Failed to hash password", nil, nil)
	}

	now := time.Now()
	user = &postgre_model.User{
		ID:        uuid.New(),
		Username:  req.Username,
		Role:      postgre_model.RoleUser,
		Status:    postgre_model.StatusActive,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = u.postgreDbRepo.CreateOneUser(ctx, user)
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}
	return pkg.NewResponse(http.StatusCreated, "User registered successfully", nil, user)
}

func (u *userAppUsecase) LoginUser(ctx context.Context, req request.UserLoginRequest) pkg.Response {
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

	user, err := u.postgreDbRepo.FetchOneUser(ctx, map[string]interface{}{"email": req.Email})
	if user == nil {
		return pkg.NewResponse(http.StatusUnauthorized, "User not found", nil, nil)
	}
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return pkg.NewResponse(http.StatusUnauthorized, "Wrong password", nil, nil)
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(config.GetJWTTTL()) * time.Minute)
	token, err := jwt_pkg.GenerateJWTTokenUser(jwt_pkg.UserJWTClaims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    "user",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiredAt),
		},
	})
	if err != nil {
		return pkg.NewResponse(http.StatusInternalServerError, err.Error(), nil, nil)
	}

	data := map[string]interface{}{"token": token, "expired_at": expiredAt, "user": user}
	return pkg.NewResponse(http.StatusOK, "Login successful", nil, data)
}
