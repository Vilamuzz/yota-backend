package jwt_pkg

import (
	"github.com/Vilamuzz/yota-backend/pkg/enum"
	"github.com/golang-jwt/jwt/v5"
)

type UserJWTClaims struct {
	AccountID  string          `json:"accountId"`
	Roles      []enum.RoleName `json:"roles"`
	ActiveRole enum.RoleName   `json:"activeRole"`
	jwt.RegisteredClaims
}
