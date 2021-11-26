package httpadminserver

import (
	"Pando/test/mock"
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

var pando, _ = mock.NewPandoMock()

func TestServerBadAddr(t *testing.T) {
	_, err := New("", pando.Registry)
	assert.Contains(t, err.Error(), "bad ingest address in config")
}

func TestServerStartShutdown(t *testing.T) {
	s, err := New("/ip4/127.0.0.1/tcp/9001", pando.Registry)
	assert.NoError(t, err)

	go func() {
		err = s.Start()
		assert.EqualError(t, err, http.ErrServerClosed.Error())
	}()
	time.Sleep(time.Millisecond * 50)
	err = s.Shutdown(context.Background())
	assert.NoError(t, err)

	// start again
	go func() {
		err = s.Start()
		assert.EqualError(t, err, http.ErrServerClosed.Error())
	}()
	time.Sleep(time.Millisecond * 50)
	err = s.Shutdown(context.Background())
}

func TestServerOption(t *testing.T) {
	_, err := New("/ip4/127.0.0.1/tcp/9001", pando.Registry, WriteTimeout(time.Second), ReadTimeout(time.Second))
	assert.NoError(t, err)
}
