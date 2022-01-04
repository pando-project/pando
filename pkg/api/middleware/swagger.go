package middleware

import (
	"github.com/gin-gonic/gin"
	swagMiddleware "github.com/go-openapi/runtime/middleware"

	"net/http"
)

type wrappedResponseWriter struct {
	gin.ResponseWriter
	writer http.ResponseWriter
}

func (w *wrappedResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

func (w *wrappedResponseWriter) WriteString(s string) (int, error) {
	return w.ResponseWriter.WriteString(s)
}

type nextRequestHandler struct {
	c *gin.Context
}

func (h *nextRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.c.Writer = &wrappedResponseWriter{h.c.Writer, w}
	h.c.Next()
}

func WithAPIDoc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		opts := swagMiddleware.RedocOpts{
			BasePath: "/swagger",
			Path:     "/doc",
			SpecURL:  "/swagger/specs",
			Title:    "Pando API Documentation",
		}
		swagMiddleware.Redoc(opts, &nextRequestHandler{ctx}).ServeHTTP(ctx.Writer, ctx.Request)
	}
}
