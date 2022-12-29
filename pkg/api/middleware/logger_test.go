package middleware

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithLoggerFormatter(t *testing.T) {
	t.Run("TestWithLoggerFormatter", func(t *testing.T) {
		logOutput := bytes.NewBuffer([]byte{})
		router := gin.New()
		router.GET("/", WithLoggerFormatter(logOutput), func(ctx *gin.Context) {
			ctx.Data(http.StatusOK, "text/plain", []byte("hello"))
		})

		server := httptest.NewServer(router)
		_, err := server.Client().Get("http://" + server.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		server.Close()

		logSlice := strings.Split(logOutput.String(), " ")
		for i, content := range logSlice {
			fmt.Printf("%d <=> %s\n", i, content)
		}

		localIP := net.ParseIP(logSlice[0])
		asserts := assert.New(t)
		asserts.NotNil(localIP)

		_, err = time.Parse(time.RFC3339, logSlice[2][1:len(logSlice[2])-1])
		asserts.Nil(err)

		asserts.Equal(http.MethodGet, logSlice[3])

		httpMajorVersion, httpMinorVersion, ok := http.ParseHTTPVersion(logSlice[5])
		asserts.Equal(true, ok)
		asserts.Equal(1, httpMajorVersion)
		asserts.Equal(1, httpMinorVersion)

		asserts.Equal(strconv.Itoa(http.StatusOK), logSlice[6])
	})
}
