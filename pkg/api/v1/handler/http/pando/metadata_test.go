package pando

import (
	"encoding/json"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando-store/pkg/types/cbortypes"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/api/types"
	"github.com/kenlabs/pando/pkg/util/cids"
	"github.com/kenlabs/pando/test/mock"
	. "github.com/smartystreets/goconvey/convey"

	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var mockAPI *API

func init() {
	log.SetAllLoggers(log.LevelDebug)
	gin.SetMode(gin.TestMode)

	var err error
	mockAPI, err = newHttpAPIMock()
	if err != nil {
		panic(err)
	}
}

func TestMetadataList(t *testing.T) {
	Convey("TestMetadataList", t, func() {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		//Convey("Given a valid cid list, should return a json with the cid list", func() {
		//	testCidListStr := []string{
		//		"bafy2bzacebxvzutul3nqhdalyxqphxyrpw2xfxa4dfuiew5uhyg2phln444us",
		//		"bafy2bzacedwt7fxhatcwqi6o7nkxyqnyunxzyijse5qgjyrjhsfock3nemae2",
		//		"bafy2bzaceabnw5lnqxytayqjqm5e5sjrlqxtht3lnitfyrv6weyup7zxw2dyc",
		//	}
		//
		//	testCidList, err := cids.DecodeAndPadSnapShotList(testCidListStr)
		//	data, err := json.Marshal(testCidList)
		//	if err != nil {
		//		t.Error(err)
		//	}
		//	patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller),
		//		"SnapShotList",
		//		func() ([]byte, error) {
		//			return data, nil
		//		})
		//	defer patch.Reset()
		//
		//	mockAPI.snapShotList(testContext)
		//	respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
		//	if err != nil {
		//		t.Error(err)
		//	}
		//
		//	resp := &types.ResponseJson{}
		//	err = json.Unmarshal(respBody, &resp)
		//	if err != nil {
		//		t.Error(err)
		//	}
		//
		//	respDataBytes, err := json.Marshal(resp.Data)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//
		//	var respCidList store.SnapShotList
		//	err = json.Unmarshal(respDataBytes, &respCidList)
		//	if err != nil {
		//		t.Error(err)
		//	}
		//
		//	So(&respCidList, ShouldResemble, testCidList)
		//	So(resp.Code, ShouldEqual, http.StatusOK)
		//	So(resp.Message, ShouldEqual, "OK")
		//})

		Convey("Given an monkey error, should return a monkey error resp", func() {
			patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller),
				"SnapShotList",
				func() ([]byte, error) {
					return nil, fmt.Errorf("monkey error")
				})
			defer patch.Reset()
			mockAPI.snapShotList(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			if err != nil {
				t.Error(err)
			}

			resp := &types.ResponseJson{}
			err = json.Unmarshal(respBody, &resp)

			So(resp.Message, ShouldEqual, "monkey error")
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
			So(resp.Data, ShouldBeNil)
		})
	})
}

func TestMetadataSnapshot(t *testing.T) {
	Convey("TestMetadataSnapshot", t, func() {
		responseRecorder := httptest.NewRecorder()
		testContext, _ := gin.CreateTestContext(responseRecorder)

		Convey("Given a non-nil snapshot, should return the snapshot", func() {
			testCidListStr := []string{
				"baguqeeqqw34gtnf4q6jtz5bgfyjnmf3jzi",
				"baguqeeqq3p7rttw3dgpahjiu53e4d6lqay",
			}
			testCidList, err := cids.DecodeCidStrList(testCidListStr)
			if err != nil {
				t.Error(err)
			}

			testSnapshot := cbortypes.SnapShot{
				Update: map[string]*cbortypes.Metalist{
					"12D3KooWSS3sEujyAXB9SWUvVtQZmxH6vTi9NitqaaRQoUjeEk3M": {testCidList},
				},
				PrevSnapShot: "bafy2bzacebxvzutul3nqhdalyxqphxyrpw2xfxa4dfuiew5uhyg2phln444us",
				Height:       1,
				CreateTime:   1646124783926371600,
			}

			testSnapshotRes, err := json.Marshal(testSnapshot)
			patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller),
				"MetadataSnapShot",
				func(_ context.Context, _ string, _ string) ([]byte, error) {
					return testSnapshotRes, nil
				},
			)
			defer patch.Reset()
			mockAPI.metadataSnapshot(testContext)

			var resp types.ResponseJson
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			if err != nil {
				t.Error(err)
			}
			if err = json.Unmarshal(respBody, &resp); err != nil {
				t.Error(err)
			}
			respData, err := json.Marshal(resp.Data)
			if err != nil {
				t.Error(err)
			}

			var actualSnapshot cbortypes.SnapShot
			err = json.Unmarshal(respData, &actualSnapshot)
			if err != nil {
				t.Error(err)
			}

			So(actualSnapshot, ShouldResemble, testSnapshot)
			So(resp.Code, ShouldEqual, http.StatusOK)
			So(resp.Message, ShouldEqual, "metadataSnapshot found")
		})

		Convey("Given an monkey error, should return a monkey error resp", func() {
			patch := gomonkey.ApplyMethodFunc(reflect.TypeOf(mockAPI.controller),
				"MetadataSnapShot",
				func(_ context.Context, _ string, _ string) ([]byte, error) {
					return nil, fmt.Errorf("monkey error")
				},
			)
			defer patch.Reset()

			mockAPI.metadataSnapshot(testContext)
			respBody, err := ioutil.ReadAll(responseRecorder.Result().Body)
			if err != nil {
				t.Error(err)
			}

			var resp types.ResponseJson
			err = json.Unmarshal(respBody, &resp)

			So(resp.Message, ShouldEqual, "monkey error")
			So(resp.Code, ShouldEqual, http.StatusBadRequest)
			So(resp.Data, ShouldBeNil)
		})
	})
}

func newHttpAPIMock() (*API, error) {
	pandoMock, err := mock.NewPandoMock()
	if err != nil {
		return nil, err
	}

	//dsLevelDB, err := leveldb.NewDatastore("/tmp/datastore", nil)
	//if err != nil {
	//	return nil, err
	//}

	apiCore := &core.Core{
		LegsCore: pandoMock.Core,
		Registry: pandoMock.Registry,
		StoreInstance: &core.StoreInstance{
			MutexDataStore: pandoMock.DS.(*sync.MutexDatastore),
			CacheStore:     pandoMock.CS,
			PandoStore:     pandoMock.PS,
		},
	}
	return NewV1HttpAPI(gin.Default(), apiCore, pandoMock.Opt), nil
}
