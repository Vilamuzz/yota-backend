package jwt_pkg

import (
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.GetJWTSecretKey()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
