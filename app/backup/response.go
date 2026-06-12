package backup

import "time"

type BackupResponse struct {
	ID              string    `json:"id"`
	Filename        string    `json:"filename"`
	Size            int64     `json:"size"`
	CreatedAt       time.Time `json:"createdAt"`
	DurationSeconds int64     `json:"durationSeconds"`
}

type BackupURLResponse struct {
	URL string `json:"url"`
}

type BackupCleanupResponse struct {
	DeletedCount  int `json:"deletedCount"`
	RetentionDays int `json:"retentionDays"`
}

func (b *Backup) toBackupResponse() BackupResponse {
	return BackupResponse{
		ID:              b.ID.String(),
		Filename:        b.Filename,
		Size:            b.Size,
		CreatedAt:       b.CreatedAt,
		DurationSeconds: b.Duration,
	}
}

func toBackupListResponse(backups []Backup) []BackupResponse {
	res := make([]BackupResponse, len(backups))
	for i, b := range backups {
		res[i] = b.toBackupResponse()
	}
	return res
}
