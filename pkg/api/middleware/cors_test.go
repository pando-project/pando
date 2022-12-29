package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithCorsAllowAllOrigin(t *testing.T) {
	t.Run("TestWithCorsAllowAllOrigin", func(t *testing.T) {
		router := gin.New()
		router.GET("/", gin.Logger(), WithCorsAllowAllOrigin(), func(ctx *gin.Context) {
			ctx.Data(http.StatusOK, "text/plain", nil)
		})

		server := httptest.NewServer(router)
		defer server.Close()

		req, _ := http.NewRequest("GET", "http://"+server.Listener.Addr().String(), nil)
		req.Header.Add("Origin", "https://kencloud.com")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	})
}
