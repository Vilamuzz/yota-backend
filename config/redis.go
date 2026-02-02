package config

import (
	"os"
	"strconv"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	Enabled  bool
}

func GetRedisConfig() RedisConfig {
	port, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if port == 0 {
		port = 6379
	}

	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

	enabled := os.Getenv("REDIS_ENABLED")
	if enabled == "" {
		enabled = "false"
	}

	return RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     port,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
		Enabled:  enabled == "true",
	}
}
