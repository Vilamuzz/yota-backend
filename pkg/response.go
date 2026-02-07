package pkg

type Response struct {
	Status     int               `json:"status"`
	Message    string            `json:"message,omitempty"`
	Validation map[string]string `json:"validation,omitempty"`
	Data       interface{}       `json:"data,omitempty"`
}

type CursorPagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	PrevCursor string `json:"prev_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	HasPrev    bool   `json:"has_prev"`
	Limit      int    `json:"limit"`
}

func NewResponse(status int, message string, validation map[string]string, data interface{}) Response {
	return Response{
		Status:     status,
		Message:    message,
		Validation: validation,
		Data:       data,
	}
}