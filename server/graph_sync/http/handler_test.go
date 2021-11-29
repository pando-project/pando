package http

import (
	"Pando/test/mock"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var pando, _ = mock.NewPandoMock()

func TestSubProvider(t *testing.T) {
	if pando == nil {
		t.Fatal("nil pando mock")
	}

	h := newHandler(pando.Core)
	p := "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"

	req := httptest.NewRequest("GET", "https://localhost:9002/graph/sub/"+p, nil)
	ctx := context.Background()
	m := map[string]string{"peerid": p}
	ctx = context.WithValue(ctx, 0, m)

	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	h.SubProvider(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Log(resp.StatusCode)
		t.Fatal("expected response to be", http.StatusOK)
	}
}
