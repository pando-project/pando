package metrics

import (
	"context"
	"github.com/filecoin-project/go-indexer-core/metrics"
	"go.opencensus.io/tag"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	logging "github.com/ipfs/go-log/v2"
	promclient "github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Measures
var (
	// pando handlers
	GetPandoSubscribeLatency = stats.Float64("get/pando/subscribe_latency",
		"Time to subscribe a provider", stats.UnitMilliseconds)
	GetPandoInfoLatency = stats.Float64("get/pando/info_latency",
		"Time to list pando info", stats.UnitMilliseconds)

	// provider handlers
	PostProviderRegisterLatency = stats.Float64("post/provider/register_latency",
		"Time to respond to register a provider", stats.UnitMilliseconds)

	// metadata handlers
	GetMetadataListLatency = stats.Float64("get/metadata/list_latency",
		"Time to fetch metadata snapshots", stats.UnitMilliseconds)
	GetMetadataSnapshotLatency = stats.Float64("get/metadata/snapshot_latency",
		"Time to fetch snapshot info", stats.UnitMilliseconds)

	// go-legs graph persistence
	GraphPersistenceLatency = stats.Float64("sync/graph/persistence_latency",
		"Time to persistence DAG", stats.UnitMilliseconds)
)

// Views
var (
	bounds       = []float64{0, 1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 300, 400, 500, 1000, 2000, 5000}
	builtinViews = []*view.View{
		{Measure: PostProviderRegisterLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetPandoSubscribeLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetMetadataListLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetMetadataSnapshotLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GraphPersistenceLatency, Aggregation: view.Distribution(bounds...)},
	}
)

func APITimer(ctx context.Context, m *stats.Float64Measure) func() {
	start := time.Now()

	return func() {
		_ = stats.RecordWithOptions(
			ctx,
			stats.WithTags(tag.Insert(metrics.Method, "api")),
			stats.WithMeasurements(m.M(coremetrics.MsecSince(start))),
		)
	}
}

var log = logging.Logger("metrics")

// Handler creates an HTTP router for serving metric info
func Handler(views []*view.View) http.Handler {
	// Register default views
	err := view.Register(builtinViews...)
	if err != nil {
		log.Errorf("cannot register metrics default views: %s", err)
	}
	// Register other views
	err = view.Register(views...)
	if err != nil {
		log.Errorf("cannot register metrics views: %s", err)
	}
	registry, ok := promclient.DefaultRegisterer.(*promclient.Registry)
	if !ok {
		log.Warnf("failed to export default prometheus registry; "+
			"some metrics will be unavailable; unexpected types: %T", promclient.DefaultRegisterer)
	}
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Registry:  registry,
		Namespace: "pando",
	})
	if err != nil {
		log.Errorf("could not create the prometheus stats exporter: %v", err)
	}

	return exporter
}
