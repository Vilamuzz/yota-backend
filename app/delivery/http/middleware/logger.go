package middleware

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *appMiddleware) Logger(writer io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			formatted, _ := json.Marshal(struct {
				Time     string `json:"time"`
				Status   int    `json:"status"`
				Method   string `json:"method"`
				Path     string `json:"path"`
				Latency  string `json:"latency"`
				ClientIP string `json:"client_ip"`
				Error    string `json:"error"`
			}{
				Time:     param.TimeStamp.Format(time.RFC3339),
				Status:   param.StatusCode,
				Method:   param.Method,
				Path:     param.Path,
				Latency:  param.Latency.String(),
				ClientIP: param.ClientIP,
				Error:    param.ErrorMessage,
			})
			return string(formatted) + "\n"
		},
		Output: writer,
	})
}
