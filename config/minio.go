package config

import (
    "os"
    "strconv"
)

type MinIOConfig struct {
    Endpoint   string
    AccessKey  string
    SecretKey  string
    BucketName string
    UseSSL     bool
    Region     string
}

func GetMinIOConfig() MinIOConfig {
    useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
    if os.Getenv("MINIO_USE_SSL") == "" {
        useSSL = false
    }

    return MinIOConfig{
        Endpoint:   os.Getenv("MINIO_ENDPOINT"),
        AccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
        SecretKey:  os.Getenv("MINIO_SECRET_KEY"),
        BucketName: os.Getenv("MINIO_BUCKET_NAME"),
        UseSSL:     useSSL,
        Region:     os.Getenv("MINIO_REGION"),
    }
}