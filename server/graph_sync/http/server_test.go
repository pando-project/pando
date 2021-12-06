package http

import (
	"Pando/test/mock"
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

func TestCreateServer(t *testing.T) {
	Convey("test create server", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		Convey("bad address", func() {
			_, err := New("12345", core)
			So(err.Error(), ShouldContainSubstring, "bad ingest address")
		})
		Convey("correct create", func() {
			s, err := New("/ip4/127.0.0.1/tcp/57654", core)
			So(err, ShouldBeNil)
			errCh := make(chan error)
			go func() {
				err = s.Start()
				errCh <- err
			}()
			time.Sleep(time.Second)
			err = s.Shutdown(context.Background())
			So(err, ShouldBeNil)
			So(<-errCh, ShouldResemble, http.ErrServerClosed)
		})
	})
}
