package log

// LogResponse is the admin-facing view of a single audit log entry.
type LogResponse struct {
	ID         string  `json:"id"`
	UserID     *string `json:"userId"`
	Action     string  `json:"action"`
	EntityType string  `json:"entityType"`
	EntityID   string  `json:"entityId"`
	OldValue   string  `json:"oldValue,omitempty"`
	NewValue   string  `json:"newValue,omitempty"`
	CreatedAt  string  `json:"createdAt"`
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
