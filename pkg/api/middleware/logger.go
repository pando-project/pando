package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func WithLoggerFormatter() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func WithLoggerToStdOut() gin.HandlerFunc {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{})

	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		latency := time.Since(startTime)
		requestMethod := ctx.Request.Method
		requestURI := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()
		bytesSent := ctx.Writer.Size()
		httpReferer := ctx.Request.Referer()
		userAgent := ctx.Request.UserAgent()
		xForwardedFor := ctx.Request.Header.Get("X-Forwarded-For")
		if xForwardedFor == "" {
			xForwardedFor = "-"
		}

		// log format
		// '$remote_addr $remote_user [$time_local] "$request" '
		// '$status $body_bytes_sent "$http_referer" '
		// '$http_user_agent $http_x_forwarded_for $request_time';
		// 110.156.114.121 - [11/Aug/2017:09:57:19 +0800] "GET /rest/mywork/latest/status/notification/count?_=1502416641768 HTTP/1.1"
		// 200 67 "http://wiki.wang-inc.com/pages/viewpage.action?pageId=11174759"
		// Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.78 Safari/537.36
		// - 0.006
		logger.Infof("%s - [%s] \"%s %s\" %d %d \"%s\" %s %s %d",
			clientIP,
			startTime,
			requestMethod,
			requestURI,
			statusCode,
			bytesSent,
			httpReferer,
			userAgent,
			xForwardedFor,
			latency,
		)
	}
}
