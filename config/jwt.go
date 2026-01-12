package config

import (
	"os"
	"strconv"
)

func GetJWTSecretKeySuperadmin() string {
	return os.Getenv("JWT_SECRET_KEY_SUPERADMIN")
}

func GetJWTSecretKeyAdmin() string {
	return os.Getenv("JWT_SECRET_KEY_ADMIN")
}

func GetJWTSecretKeyUser() string {
	return os.Getenv("JWT_SECRET_KEY_USER")
}

func GetJWTTTL() int {
	ttl, _ := strconv.Atoi(os.Getenv("JWT_TTL"))
	if ttl == 0 {
		ttl = 60 //default value 60 minutes
	}
	return ttl
}
