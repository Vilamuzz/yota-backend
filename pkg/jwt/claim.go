package jwt_pkg

import "github.com/golang-jwt/jwt/v5"

type UserJWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}
