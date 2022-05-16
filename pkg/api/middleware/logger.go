package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

func WithLoggerFormatter(w ...io.Writer) gin.HandlerFunc {
	logFormatter := func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] %s %s %s %d %s %s %s\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}

	if w == nil {
		w = []io.Writer{gin.DefaultWriter}
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: logFormatter,
		Output:    w[0],
	})
}
