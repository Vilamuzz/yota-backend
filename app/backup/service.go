package backup

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service interface {
	CreateBackup(ctx context.Context) (*BackupMetadata, error)
	ListBackups(ctx context.Context) ([]BackupMetadata, error)
	GetBackupURL(ctx context.Context, backupID string) (string, error)
	DeleteBackup(ctx context.Context, backupID string) error
	CleanupOldBackups(ctx context.Context, retentionDays int) (int, error)
}

type service struct {
	db          *gorm.DB
	minioClient *minio.Client
	bucketName  string
	timeout     time.Duration
}

func NewService(db *gorm.DB, minioClient *minio.Client) Service {
	bucketName := os.Getenv("RUSTFS_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "default-bucket"
	}

	return &service{
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
		timeout:     30 * time.Second,
	}
}

func (s *service) CreateBackup(ctx context.Context) (*BackupMetadata, error) {
	startTime := time.Now()
	backupID := uuid.New().String()
	filename := fmt.Sprintf("backup_%s_%d.sql", time.Now().Format("20060102_150405"), time.Now().Unix())
	tempFile := filepath.Join(os.TempDir(), filename)

	logrus.Infof("Backup: starting backup [%s] -> %s", backupID, filename)

	connStr := s.extractConnectionString()
	if connStr == "" {
		logrus.Errorf("Backup: failed to extract connection string")
		return nil, fmt.Errorf("failed to extract database connection string")
	}

	// Execute pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump", connStr)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Backup: pg_dump failed: %v, stderr: %s", err, errOut.String())
		return nil, fmt.Errorf("pg_dump failed: %w", err)
	}

	// Write to temp file
	if err := os.WriteFile(tempFile, out.Bytes(), 0600); err != nil {
		logrus.Errorf("Backup: failed to write temp file: %v", err)
		return nil, fmt.Errorf("failed to write backup file: %w", err)
	}
	defer os.Remove(tempFile)

	// Upload to S3
	fileInfo, err := os.Stat(tempFile)
	if err != nil {
		logrus.Errorf("Backup: failed to stat file: %v", err)
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	file, err := os.Open(tempFile)
	if err != nil {
		logrus.Errorf("Backup: failed to open file for upload: %v", err)
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	backupPath := fmt.Sprintf("backups/%s", filename)
	_, err = s.minioClient.PutObject(ctx, s.bucketName, backupPath, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/sql",
	})
	if err != nil {
		logrus.Errorf("Backup: failed to upload to S3: %v", err)
		return nil, fmt.Errorf("failed to upload backup to S3: %w", err)
	}

	duration := time.Since(startTime).Seconds()
	metadata := &BackupMetadata{
		ID:        backupID,
		Filename:  filename,
		Size:      fileInfo.Size(),
		CreatedAt: time.Now(),
		Duration:  int64(duration),
	}

	logrus.Infof("Backup: successfully created backup [%s] (size: %d bytes, duration: %.2fs)", backupID, fileInfo.Size(), duration)
	return metadata, nil
}

func (s *service) ListBackups(ctx context.Context) ([]BackupMetadata, error) {
	listChan := s.minioClient.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "backups/",
		Recursive: true,
	})

	backups := make([]BackupMetadata, 0)
	for obj := range listChan {
		if obj.Err != nil {
			logrus.Errorf("Backup: error listing objects: %v", obj.Err)
			return nil, fmt.Errorf("failed to list backups: %w", obj.Err)
		}

		backup := BackupMetadata{
			ID:        uuid.New().String(),
			Filename:  filepath.Base(obj.Key),
			Size:      obj.Size,
			CreatedAt: obj.LastModified,
		}
		backups = append(backups, backup)
	}

	logrus.Infof("Backup: listed %d backups", len(backups))
	return backups, nil
}

func (s *service) GetBackupURL(ctx context.Context, backupID string) (string, error) {
	logrus.Infof("Backup: getting URL for backup [%s]", backupID)
	return "", fmt.Errorf("backup URL retrieval not fully implemented - requires DB mapping")
}

func (s *service) DeleteBackup(ctx context.Context, backupID string) error {
	logrus.Infof("Backup: deleting backup [%s]", backupID)
	return fmt.Errorf("backup deletion not fully implemented - requires DB mapping")
}

func (s *service) CleanupOldBackups(ctx context.Context, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		retentionDays = 7
	}

	logrus.Infof("Backup: cleanup started - removing backups older than %d days", retentionDays)
	
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	deletedCount := 0

	listChan := s.minioClient.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "backups/",
		Recursive: true,
	})

	for obj := range listChan {
		if obj.Err != nil {
			logrus.Errorf("Backup: error listing objects during cleanup: %v", obj.Err)
			continue
		}

		if obj.LastModified.Before(cutoffTime) {
			err := s.minioClient.RemoveObject(ctx, s.bucketName, obj.Key, minio.RemoveObjectOptions{})
			if err != nil {
				logrus.Errorf("Backup: failed to delete old backup [%s]: %v", obj.Key, err)
				continue
			}

			logrus.Infof("Backup: deleted old backup [%s] (created: %s)", obj.Key, obj.LastModified.Format(time.RFC3339))
			deletedCount++
		}
	}

	logrus.Infof("Backup: cleanup completed - deleted %d old backups", deletedCount)
	return deletedCount, nil
}

func (s *service) extractConnectionString() string {
	dbConn := os.Getenv("DB")
	if dbConn != "" {
		return dbConn
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if host != "" && user != "" && dbname != "" {
		if port == "" {
			port = "5432"
		}
		if password != "" {
			return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
		}
		return fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=disable", user, host, port, dbname)
	}

	return ""
}
