package config

import (
	"os"
	"strconv"
)

func GetJWTSecretKey() string {
	return os.Getenv("JWT_SECRET_KEY")
}

func GetJWTTTL() int {
	ttl, _ := strconv.Atoi(os.Getenv("JWT_TTL"))
	if ttl == 0 {
		ttl = 60 //default value 60 minutes
	}
	return ttl
}
