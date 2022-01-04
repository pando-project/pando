package http

import (
	"net/http"

	coreMetrics "github.com/filecoin-project/go-indexer-core/metrics"
	adapter "github.com/gwatts/gin-adapter"

	"pando/pkg/metrics"
)

func (a *API) registerMetrics() {
	a.router.GET("/metrics", adapter.Wrap(
		func(h http.Handler) http.Handler {
			return metrics.Handler(coreMetrics.DefaultViews)
		}),
	)
}
