package log

// LogQueryParams holds query params for the admin log list endpoint.
type LogQueryParams struct {
	EntityType string `form:"entityType"`
	EntityID   string `form:"entityId"`
	UserID     string `form:"userId"`
	Action     string `form:"action"`
	Limit      int    `form:"limit"`
	NextCursor string `form:"nextCursor"`
}
