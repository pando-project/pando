package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
)

func TestHttpResponse(router *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) bool) bool {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if !f(w) {
		return false
	} else {
		return true
	}
}
