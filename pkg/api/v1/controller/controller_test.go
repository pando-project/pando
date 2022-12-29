package controller

import (
	"testing"

	"github.com/pando-project/pando/pkg/api/core"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/test/mock"

	"github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-log/v2"
	"github.com/stretchr/testify/assert"
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
	t.Run("TestNewController", func(t *testing.T) {
		t.Run("NewController should return a controller with non-nil opt and core", func(t *testing.T) {
			c := &core.Core{}
			opt := &option.DaemonOptions{}
			controller := New(c, opt)

			asserts := assert.New(t)
			asserts.Equal(c, controller.Core)
			asserts.Equal(opt, controller.Options)
		})
	})
}

func newMockController() (*Controller, error) {
	pandoMock, err := mock.NewPandoMock()
	if err != nil {
		return nil, err
	}

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
