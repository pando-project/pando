package httpclient

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
	"time"
)

func TestCreateClient(t *testing.T) {
	Convey("test create http client", t, func() {
		testCase := []struct {
			baseUrl     string
			resource    string
			defaultPort int
			options     []Option
		}{
			{
				baseUrl:     "dsad",
				resource:    "dasda",
				defaultPort: 123,
				options:     []Option{Timeout(time.Second)},
			},
		}
		for _, tc := range testCase {
			u, _, e := New(tc.baseUrl, tc.resource, tc.defaultPort, tc.options...)
			So(e, ShouldBeNil)
			So(u.String(), ShouldEqual, "http://"+tc.baseUrl+":"+strconv.Itoa(tc.defaultPort)+"/"+tc.resource)
		}
	})
}

func TestReadError(t *testing.T) {
	Convey("test read error", t, func() {
		errorBytes := []byte("error request")
		statusCode := 400
		err := ReadError(statusCode, errorBytes)
		So(err, ShouldResemble, errors.New("400 Bad Request: error request"))
	})
}
