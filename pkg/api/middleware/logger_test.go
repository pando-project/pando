package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestWithLoggerFormatter(t *testing.T) {
	Convey("TestWithLoggerFormatter", t, func() {
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
		So(localIP, ShouldNotBeNil)

		_, err = time.Parse(time.RFC3339, logSlice[2][1:len(logSlice[2])-1])
		So(err, ShouldBeNil)

		So(logSlice[3], ShouldEqual, http.MethodGet)

		httpMajorVersion, httpMinorVersion, ok := http.ParseHTTPVersion(logSlice[5])
		So(ok, ShouldEqual, true)
		So(httpMajorVersion, ShouldEqual, 1)
		So(httpMinorVersion, ShouldEqual, 1)

		So(logSlice[6], ShouldEqual, strconv.Itoa(http.StatusOK))
	})
}
