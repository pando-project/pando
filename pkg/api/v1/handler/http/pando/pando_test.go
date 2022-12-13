package pando

import (
	"encoding/json"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/pando-project/pando/pkg/api/types"
	"github.com/pando-project/pando/pkg/api/v1/model"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPandoInfo(t *testing.T) {
	Convey("TestPandoInfo", t, func() {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		Convey("Given valid pando info without error, should return the pando info resp", func() {
			testPandoInfo := model.PandoInfo{
				PeerID: "12D3KooWDhanS6yHjR4CjbtnRtrMFgbzb3YZLGAqn87m442MpEEK",
				Addresses: model.APIAddresses{
					HttpAPI:      "/ip4/127.0.0.1/tcp/9001",
					GraphQLAPI:   "/ip4/127.0.0.1/tcp/9002",
					GraphSyncAPI: "/ip4/127.0.0.1/tcp/9003",
				},
			}

			patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller), "PandoInfo",
				func() (*model.PandoInfo, error) {
					return &testPandoInfo, nil
				})
			defer patch.Reset()

			mockAPI.pandoInfo(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			if err != nil {
				t.Error(err)
			}
			var resp types.ResponseJson
			if err = json.Unmarshal(respBody, &resp); err != nil {
				t.Error(err)
			}

			var actualPandoInfo model.PandoInfo
			respData, err := json.Marshal(resp.Data)
			if err != nil {
				t.Error(err)
			}
			if err = json.Unmarshal(respData, &actualPandoInfo); err != nil {
				t.Errorf("unmarshal pandoInfoData failed, err: %v", err)
			}

			So(actualPandoInfo, ShouldResemble, testPandoInfo)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Message, ShouldEqual, "OK")
		})

		Convey("Given a monkey error, should return a monkey error resp", func() {
			patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller), "PandoInfo",
				func() (*model.PandoInfo, error) {
					return nil, fmt.Errorf("monkey error")
				})
			defer patch.Reset()

			mockAPI.pandoInfo(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)

			var resp types.ResponseJson
			if err = json.Unmarshal(respBody, &resp); err != nil {
				t.Error(err)
			}

			So(resp.Message, ShouldEqual, "monkey error")
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
			So(resp.Data, ShouldBeNil)
		})
	})
}
