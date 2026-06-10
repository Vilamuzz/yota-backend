package backup

import "time"

type BackupMetadata struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Duration  int64     `json:"duration_seconds"`
}

type BackupRequest struct {
	Comment string `json:"comment" binding:"omitempty"`
}

type BackupResponse struct {
	Success   bool            `json:"success"`
	Message   string          `json:"message"`
	Backup    *BackupMetadata `json:"backup,omitempty"`
	Error     string          `json:"error,omitempty"`
}

type BackupListResponse struct {
	Success  bool              `json:"success"`
	Backups  []BackupMetadata  `json:"backups"`
	Total    int               `json:"total"`
	Error    string            `json:"error,omitempty"`
}
