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

var Method, _ = tag.NewKey("method")

// Measures
var (
	RegisterProviderLatency = stats.Float64("providers/register_latency",
		"Time to respond to register a provider", stats.UnitMilliseconds)
	SubProviderLatency = stats.Float64("graph/sub_latency",
		"Time to subscribe a provider", stats.UnitMilliseconds)
	ListMetadataLatency = stats.Float64("meta/list_latency",
		"Time to fetch metadata snapshots", stats.UnitMilliseconds)
	ListPandoInfoLatency = stats.Float64("meta/list_info",
		"Time to list pando info", stats.UnitMilliseconds)
	ListSnapshotInfoLatency = stats.Float64("meta/snap_info_latency",
		"Time to fetch snapshot info", stats.UnitMilliseconds)
	GetSnapshotByHeightLatency = stats.Float64("meta/snap_info_by_height",
		"Time to fetch snapshot info by height", stats.UnitMilliseconds)
	GraphPersistenceLatency = stats.Float64("graph/persistence_latency",
		"Time to persistence DAG", stats.UnitMilliseconds)
)

// Views
var (
	bounds       = []float64{0, 1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 300, 400, 500, 1000, 2000, 5000}
	builtinViews = []*view.View{
		{Measure: RegisterProviderLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: SubProviderLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: ListMetadataLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: ListSnapshotInfoLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetSnapshotByHeightLatency, Aggregation: view.Distribution(bounds...)},
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
			"some metrics will be unavailable; unexpected type: %T", promclient.DefaultRegisterer)
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
