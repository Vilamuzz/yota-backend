package pkg

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

type CursorData struct {
	CreatedAt time.Time
	ID        string
}

func EncodeCursor(createdAt time.Time, id string) string {
	cursorStr := fmt.Sprintf("%d|%s", createdAt.Unix(), id)
	return base64.URLEncoding.EncodeToString([]byte(cursorStr))
}

func DecodeCursor(cursor string) (*CursorData, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	var timestamp int64
	fmt.Sscanf(parts[0], "%d", &timestamp)

	return &CursorData{
		CreatedAt: time.Unix(timestamp, 0),
		ID:        parts[1],
	}, nil
}
