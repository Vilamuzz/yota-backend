package jwt_pkg

import (
	"github.com/Vilamuzz/yota-backend/config"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTTokenSuperadmin(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.GetJWTSecretKeySuperadmin()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateJWTTokenAdmin(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.GetJWTSecretKeyAdmin()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateJWTTokenUser(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(config.GetJWTSecretKeyUser()))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
