package controller

import (
	"github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/test/mock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

var mockController *Controller

func init() {
	log.SetAllLoggers(log.LevelDebug)

	var err error
	mockController, err = newMockController()
	if err != nil {
		panic(err)
	}
}

func TestNewController(t *testing.T) {
	Convey("TestNewController", t, func() {
		Convey("NewController should return a controller with non-nil opt and core", func() {
			c := &core.Core{}
			opt := &option.DaemonOptions{}
			controller := New(c, opt)

			So(controller.Core, ShouldEqual, c)
			So(controller.Options, ShouldEqual, opt)
		})
	})
}

func newMockController() (*Controller, error) {
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
	return New(apiCore, pandoMock.Opt), nil
}
