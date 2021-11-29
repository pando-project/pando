package http

import (
	"Pando/legs"
	"Pando/test/mock"
	"bytes"
	"context"
	"errors"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/gorilla/mux"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSubProvider(t *testing.T) {
	Convey("test subscribe provider", t, func() {
		// mock functions dep
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		h := newHandler(pando.Core)

		request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1", bytes.NewReader([]byte("")))
		httpWriter := httptest.NewRecorder()

		Convey("return 200 when subscribe right", func() {
			patch1 := ApplyMethod(reflect.TypeOf(pando.Core), "Subscribe",
				func(_ *legs.Core, _ context.Context, _ peer.ID) error {
					return nil
				})
			defer patch1.Reset()
			patch2 := ApplyFunc(mux.Vars, func(r *http.Request) map[string]string {
				return map[string]string{"peerid": "12D3KooWJfFoQ2D1nukmG84DEh6gGEEE49yG6rPCdHoCqhF7YyL1"}
			})
			defer patch2.Reset()
			h.SubProvider(httpWriter, request)
			So(httpWriter.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
		Convey("return 500 when subscribe failed", func() {
			patch1 := ApplyMethod(reflect.TypeOf(pando.Core), "Subscribe",
				func(_ *legs.Core, _ context.Context, _ peer.ID) error {
					return errors.New("testerror")
				})
			defer patch1.Reset()
			patch2 := ApplyFunc(mux.Vars, func(r *http.Request) map[string]string {
				return map[string]string{"peerid": "12D3KooWJfFoQ2D1nukmG84DEh6gGEEE49yG6rPCdHoCqhF7YyL1"}
			})
			defer patch2.Reset()

			h.SubProvider(httpWriter, request)
			So(httpWriter.Result().StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("return 400 when decode peerid failed", func() {
			patch := ApplyFunc(mux.Vars, func(r *http.Request) map[string]string {
				return map[string]string{"peerid": ""}
			})
			defer patch.Reset()

			h.SubProvider(httpWriter, request)
			So(httpWriter.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
		})

	})
}
