package backup

type BackupRequest struct {
	Comment string `json:"comment" form:"comment"`
}

type CleanupQueryParams struct {
	RetentionDays *int `form:"retention"`
}
