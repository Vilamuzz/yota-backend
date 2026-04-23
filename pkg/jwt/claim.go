package jwt_pkg

import (
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/golang-jwt/jwt/v5"
)

type UserJWTClaims struct {
	AccountID  string          `json:"account_id"`
	Roles      []enum.RoleName `json:"roles"`
	ActiveRole enum.RoleName   `json:"active_role"`
	jwt.RegisteredClaims
}
