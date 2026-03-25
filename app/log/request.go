package log

// LogQueryParams holds query params for the admin log list endpoint.
type LogQueryParams struct {
	EntityType string `form:"entity_type"`
	EntityID   string `form:"entity_id"`
	UserID     string `form:"user_id"`
	Action     string `form:"action"`
	Limit      int    `form:"limit"`
	NextCursor string `form:"next_cursor"`
}
