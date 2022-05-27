package pando

import (
	"bytes"
	goContext "context"
	"encoding/json"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestProviderRegister(t *testing.T) {
	Convey("TestProviderRegister", t, func() {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		Convey("When controller.ProviderRegister return nil error, should return success resp", func() {
			patch := gomonkey.ApplyMethodFunc(
				reflect.TypeOf(mockAPI.controller),
				"ProviderRegister",
				func(_ goContext.Context, _ []byte) error {
					return nil
				},
			)
			defer patch.Reset()

			req, err := http.NewRequest("POST", "http://127.0.0.1", bytes.NewBufferString("test body"))
			testContext.Request = req
			if err != nil {
				t.Error(err)
			}
			mockAPI.providerRegister(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			if err != nil {
				t.Error(err)
			}

			var resp types.ResponseJson
			if err = json.Unmarshal(respBody, &resp); err != nil {
				t.Error(err)
			}

			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Message, ShouldEqual, "register success")
		})

		Convey("When controller.ProviderRegister return an error, should return an error resp", func() {
			patch := gomonkey.ApplyMethodFunc(
				reflect.TypeOf(mockAPI.controller),
				"ProviderRegister",
				func(_ goContext.Context, _ []byte) error {
					return fmt.Errorf("monkey error")
				},
			)
			defer patch.Reset()

			req, err := http.NewRequest("POST", "http://127.0.0.1", bytes.NewBufferString("test body"))
			testContext.Request = req
			if err != nil {
				t.Error(err)
			}

			mockAPI.providerRegister(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)

			var resp types.ResponseJson
			if err = json.Unmarshal(respBody, &resp); err != nil {
				t.Error(err)
			}

			So(resp.Code, ShouldEqual, http.StatusBadRequest)
			So(resp.Message, ShouldEqual, "monkey error")
		})
	})
}
