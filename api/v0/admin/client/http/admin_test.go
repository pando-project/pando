package adminhttpclient

import (
	"Pando/test/mock"
	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"reflect"

	"context"
	"testing"
)

func TestCreateAndRegister(t *testing.T) {
	Convey("test create client and register", t, func() {
		Convey("bad url", func() {
			url := "1`2`2`.`1`3`21/**&*()////foo?query=http://bad"
			_, err := New(url)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})
		Convey("right url and register", func() {
			c, err := New("http://123.321.0.1")
			So(err, ShouldBeNil)
			addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
			peerID, pk, err := mock.GetPrivkyAndPeerID()
			So(err, ShouldBeNil)

			patch := ApplyPrivateMethod(reflect.TypeOf(http.DefaultClient), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
				res := httptest.NewRecorder()
				res.WriteHeader(200)
				return res.Result(), nil
			})
			defer patch.Reset()
			//ApplyFunc()
			err = c.Register(context.Background(), peerID, pk, addrs, "")
			So(err, ShouldBeNil)
		})
		Convey("failed http request", func() {
			c, err := New("http://123.321.0.1")
			So(err, ShouldBeNil)
			addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
			peerID, pk, err := mock.GetPrivkyAndPeerID()
			So(err, ShouldBeNil)
			patch := ApplyPrivateMethod(reflect.TypeOf(http.DefaultClient), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
				res := httptest.NewRecorder()
				res.WriteHeader(404)
				return res.Result(), http.ErrHandlerTimeout
			})
			defer patch.Reset()
			err = c.Register(context.Background(), peerID, pk, addrs, "")
			So(err, ShouldResemble, http.ErrHandlerTimeout)

		})
	})
}
