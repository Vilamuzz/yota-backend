package log

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Vilamuzz/yota-backend/pkg"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Service interface {
	// CreateLog records an admin/user action. Pass nil for userID when actor is unknown.
	CreateLog(ctx context.Context, userID *string, action, entityType, entityID string, oldVal, newVal interface{})
	ListLogs(ctx context.Context, params LogQueryParams) pkg.Response
}

type service struct {
	repo    Repository
	timeout time.Duration
}

func NewService(repo Repository, timeout time.Duration) Service {
	return &service{repo: repo, timeout: timeout}
}

func (s *service) CreateLog(ctx context.Context, userID *string, action, entityType, entityID string, oldVal, newVal interface{}) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	oldStr := marshalJSON(oldVal)
	newStr := marshalJSON(newVal)

	entry := &Log{
		ID:         uuid.New().String(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldStr,
		NewValue:   newStr,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		logrus.WithFields(logrus.Fields{
			"component":   "log.service",
			"action":      action,
			"entity_type": entityType,
			"entity_id":   entityID,
		}).WithError(err).Warn("failed to write audit log")
	}
}

func (s *service) ListLogs(ctx context.Context, params LogQueryParams) pkg.Response {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	options := map[string]interface{}{
		"limit": params.Limit,
	}
	if params.EntityType != "" {
		options["entity_type"] = params.EntityType
	}
	if params.EntityID != "" {
		options["entity_id"] = params.EntityID
	}
	if params.UserID != "" {
		options["user_id"] = params.UserID
	}
	if params.Action != "" {
		options["action"] = params.Action
	}
	if params.NextCursor != "" {
		options["next_cursor"] = params.NextCursor
	}

	logs, err := s.repo.FindAll(ctx, options)
	if err != nil {
		return pkg.NewResponse(500, "Failed to fetch logs", nil, nil)
	}

	hasMore := len(logs) > params.Limit
	if hasMore {
		logs = logs[:params.Limit]
	}

	var nextCursor string
	if hasMore && len(logs) > 0 {
		last := logs[len(logs)-1]
		nextCursor = pkg.EncodeCursor(last.CreatedAt, last.ID)
	}

	responses := make([]LogResponse, len(logs))
	for i, l := range logs {
		lCopy := l
		responses[i] = toLogResponse(&lCopy)
	}

	return pkg.NewResponse(200, "Success", nil, map[string]interface{}{
		"logs": responses,
		"pagination": pkg.CursorPagination{
			NextCursor: nextCursor,
			Limit:      params.Limit,
		},
	})
}

func marshalJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
