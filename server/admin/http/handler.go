package httpadminserver

import (
	"Pando/internal/handler"
	"Pando/internal/httpserver"
	"Pando/internal/metrics"
	"Pando/internal/registry"
	"context"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"io"
	"net/http"
	"time"
)

type httpHandler struct {
	adminHandler *handler.AdminHandler
}

func newHandler(registry *registry.Registry) *httpHandler {
	return &httpHandler{
		adminHandler: handler.NewAdminHandler(registry),
	}
}

// POST /providers
func (h *httpHandler) RegisterProvider(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorw("failed reading body", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.adminHandler.RegisterProvider(body)
	if err != nil {
		httpserver.HandleError(w, err, "register")
		log.Warnf("register failed: %s", err.Error())
		return
	}

	_ = stats.RecordWithOptions(context.Background(),
		stats.WithTags(tag.Insert(metrics.Method, "api")),
		stats.WithMeasurements(metrics.RegisterProviderLatency.M(coremetrics.MsecSince(startTime))))
	w.WriteHeader(http.StatusOK)
}
