package log

// LogResponse is the admin-facing view of a single audit log entry.
type LogResponse struct {
	ID         string  `json:"id"`
	UserID     *string `json:"user_id"`
	Action     string  `json:"action"`
	EntityType string  `json:"entity_type"`
	EntityID   string  `json:"entity_id"`
	OldValue   string  `json:"old_value,omitempty"`
	NewValue   string  `json:"new_value,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func toLogResponse(l *Log) LogResponse {
	return LogResponse{
		ID:         l.ID,
		UserID:     l.UserID,
		Action:     l.Action,
		EntityType: l.EntityType,
		EntityID:   l.EntityID,
		OldValue:   l.OldValue,
		NewValue:   l.NewValue,
		CreatedAt:  l.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
