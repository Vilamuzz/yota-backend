package pkg

type Response struct {
	Status     int               `json:"status"`
	Message    string            `json:"message,omitempty"`
	Validation map[string]string `json:"validation,omitempty"`
	Data       interface{}       `json:"data,omitempty"`
}

func NewResponse(status int, message string, validation map[string]string, data interface{}) Response {
	return Response{
		Status:     status,
		Message:    message,
		Validation: validation,
		Data:       data,
	}
}