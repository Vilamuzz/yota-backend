package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type service struct {
	repo        Repository
	minioClient *minio.Client
	bucketName  string
	timeout     time.Duration
}

type Service interface {
	CreateBackup(ctx context.Context) pkg.Response
	ListBackups(ctx context.Context) pkg.Response
	GetBackupURL(ctx context.Context, backupID string) pkg.Response
	DeleteBackup(ctx context.Context, backupID string) pkg.Response
	CleanupOldBackups(ctx context.Context, retentionDays int) pkg.Response
}

func NewService(repo Repository, minioClient *minio.Client, timeout time.Duration) Service {
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "default-bucket"
	}

	return &service{
		repo:        repo,
		minioClient: minioClient,
		bucketName:  bucketName,
		timeout:     timeout,
	}
}

func (s *service) CreateBackup(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	startTime := time.Now()
	backupID := uuid.New()
	filename := fmt.Sprintf("backup_%s_%d.sql", time.Now().Format("20060102_150405"), time.Now().Unix())
	tempFile := filepath.Join(os.TempDir(), filename)

	logrus.WithFields(logrus.Fields{
		"component": "backup.service",
		"backup_id": backupID,
		"filename":  filename,
	}).Info("starting backup execution")

	connStr := s.extractConnectionString()
	if connStr == "" {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
		}).Error("failed to extract connection string")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengekstrak string koneksi database", nil, nil)
	}

	// Execute pg_dump
	cmd := exec.CommandContext(ctx, "pg_dump", connStr)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	if err := cmd.Run(); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"stderr":    errOut.String(),
		}).WithError(err).Error("pg_dump command execution failed")
		return pkg.NewResponse(http.StatusInternalServerError, "Eksekusi pg_dump gagal", nil, nil)
	}

	// Write to temp file
	if err := os.WriteFile(tempFile, out.Bytes(), 0600); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"temp_file": tempFile,
		}).WithError(err).Error("failed to write temp backup file")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menulis file backup sementara", nil, nil)
	}
	defer os.Remove(tempFile)

	// Upload to S3
	fileInfo, err := os.Stat(tempFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"temp_file": tempFile,
		}).WithError(err).Error("failed to stat temp backup file")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membaca detail file backup", nil, nil)
	}

	file, err := os.Open(tempFile)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"temp_file": tempFile,
		}).WithError(err).Error("failed to open temp backup file for upload")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuka file backup untuk diunggah", nil, nil)
	}
	defer file.Close()

	backupPath := fmt.Sprintf("backups/%s", filename)
	_, err = s.minioClient.PutObject(ctx, s.bucketName, backupPath, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/sql",
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"s3_path":   backupPath,
		}).WithError(err).Error("failed to upload backup file to S3")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengunggah file backup ke S3", nil, nil)
	}

	duration := int64(time.Since(startTime).Seconds())
	backupRecord := &Backup{
		ID:        backupID,
		Filename:  filename,
		Size:      fileInfo.Size(),
		Duration:  duration,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateBackup(ctx, backupRecord); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"backup_id": backupID,
		}).WithError(err).Error("failed to save backup metadata to database")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menyimpan metadata backup ke database", nil, nil)
	}

	logrus.WithFields(logrus.Fields{
		"component": "backup.service",
		"backup_id": backupID,
		"size":      fileInfo.Size(),
		"duration":  duration,
	}).Info("backup successfully created and metadata stored")

	return pkg.NewResponse(http.StatusCreated, "Backup berhasil dibuat", nil, backupRecord.toBackupResponse())
}

func (s *service) ListBackups(ctx context.Context) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	backups, err := s.repo.FindAllBackups(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
		}).WithError(err).Error("failed to retrieve backups from database")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mengambil data backup", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil mengambil daftar backup", nil, toBackupListResponse(backups))
}

func (s *service) GetBackupURL(ctx context.Context, backupID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(backupID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Format ID tidak valid", map[string]string{"id": "Format ID backup tidak valid"}, nil)
	}

	backup, err := s.repo.FindOneBackup(ctx, backupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Backup tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"backup_id": backupID,
		}).WithError(err).Error("failed to find backup record")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan data backup", nil, nil)
	}

	backupPath := fmt.Sprintf("backups/%s", backup.Filename)
	presignedURL, err := s.minioClient.PresignedGetObject(ctx, s.bucketName, backupPath, 1*time.Hour, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"s3_path":   backupPath,
		}).WithError(err).Error("failed to generate presigned S3 url")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal membuat tautan unduhan backup", nil, nil)
	}

	return pkg.NewResponse(http.StatusOK, "Berhasil menghasilkan tautan unduhan backup", nil, BackupURLResponse{
		URL: presignedURL.String(),
	})
}

func (s *service) DeleteBackup(ctx context.Context, backupID string) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if _, err := uuid.Parse(backupID); err != nil {
		return pkg.NewResponse(http.StatusBadRequest, "Format ID tidak valid", map[string]string{"id": "Format ID backup tidak valid"}, nil)
	}

	backup, err := s.repo.FindOneBackup(ctx, backupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkg.NewResponse(http.StatusNotFound, "Backup tidak ditemukan", nil, nil)
		}
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"backup_id": backupID,
		}).WithError(err).Error("failed to find backup record for deletion")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menemukan data backup", nil, nil)
	}

	backupPath := fmt.Sprintf("backups/%s", backup.Filename)
	err = s.minioClient.RemoveObject(ctx, s.bucketName, backupPath, minio.RemoveObjectOptions{})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"s3_path":   backupPath,
		}).WithError(err).Error("failed to delete backup from S3")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus file backup dari S3", nil, nil)
	}

	if err := s.repo.DeleteBackup(ctx, backupID); err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"backup_id": backupID,
		}).WithError(err).Error("failed to soft delete backup record in database")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal menghapus metadata backup dari database", nil, nil)
	}

	logrus.WithFields(logrus.Fields{
		"component": "backup.service",
		"backup_id": backupID,
	}).Info("backup successfully deleted from S3 and database")

	return pkg.NewResponse(http.StatusOK, "Backup berhasil dihapus", nil, nil)
}

func (s *service) CleanupOldBackups(ctx context.Context, retentionDays int) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if retentionDays <= 0 {
		retentionDays = 7
	}

	logrus.WithFields(logrus.Fields{
		"component":      "backup.service",
		"retention_days": retentionDays,
	}).Info("starting manual backup cleanup")

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	backups, err := s.repo.GetOldBackups(ctx, cutoffTime)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
		}).WithError(err).Error("failed to query old backups for cleanup")
		return pkg.NewResponse(http.StatusInternalServerError, "Gagal mencari file backup lama", nil, nil)
	}

	deletedCount := 0
	for _, b := range backups {
		backupPath := fmt.Sprintf("backups/%s", b.Filename)
		err := s.minioClient.RemoveObject(ctx, s.bucketName, backupPath, minio.RemoveObjectOptions{})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "backup.service",
				"s3_path":   backupPath,
			}).WithError(err).Warn("failed to delete old backup from S3 during cleanup")
			continue
		}

		if err := s.repo.DeleteBackup(ctx, b.ID.String()); err != nil {
			logrus.WithFields(logrus.Fields{
				"component": "backup.service",
				"backup_id": b.ID,
			}).WithError(err).Warn("failed to soft delete old backup metadata from database during cleanup")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"component": "backup.service",
			"backup_id": b.ID,
			"filename":  b.Filename,
		}).Info("cleaned up old backup file")
		deletedCount++
	}

	logrus.WithFields(logrus.Fields{
		"component":     "backup.service",
		"deleted_count": deletedCount,
	}).Info("cleanup completed")

	return pkg.NewResponse(http.StatusOK, "Pembersihan backup lama berhasil diselesaikan", nil, BackupCleanupResponse{
		DeletedCount:  deletedCount,
		RetentionDays: retentionDays,
	})
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
