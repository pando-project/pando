package metrics

import (
	"context"
	"github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/kenlabs/pando/pkg/util/log"
	"go.opencensus.io/tag"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
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
	GetRegisteredProviderInfoLatency = stats.Float64("get/provider/registered_info_latency",
		"Time to respond to get registered provider(s) info", stats.UnitMilliseconds)

	GetProviderHeadLatency = stats.Float64("get/provider/provider_head_latency",
		"Time to respond to get provider's head", stats.UnitMilliseconds)

	// metadata handlers
	GetMetadataListLatency = stats.Float64("get/metadata/list_latency",
		"Time to fetch metadata snapshots", stats.UnitMilliseconds)
	GetMetadataSnapshotLatency = stats.Float64("get/metadata/snapshot_latency",
		"Time to fetch snapshot info", stats.UnitMilliseconds)
	GetMetadataInclusionLatency = stats.Float64("get/metadata/inclusion_latency",
		"Time to fetch meta inclusion", stats.UnitMilliseconds)
	PostMetadataQueryLatency = stats.Float64("post/metadata/query_latency",
		"Time to query metadata", stats.UnitMilliseconds)

	// go-legs graph persistence
	GraphPersistenceLatency = stats.Float64("sync/graph/persistence_latency",
		"Time to persistence DAG", stats.UnitMilliseconds)

	// notifications count updated by provider
	ProviderNotificationCount = stats.Int64("sync/notification/count",
		"Provider notifications count", stats.UnitDimensionless)

	// payload count received from provider
	ProviderPayloadCount = stats.Int64("sync/payload/count",
		"Provider payload count", stats.UnitDimensionless)
)

// Views
var (
	providerTagKey, _ = tag.NewKey("provider")
	bounds            = []float64{0, 1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 300, 400, 500, 1000, 2000, 5000}
	builtinViews      = []*view.View{
		{Measure: PostProviderRegisterLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetProviderHeadLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetRegisteredProviderInfoLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetPandoSubscribeLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetMetadataListLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetMetadataSnapshotLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GetMetadataInclusionLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: PostMetadataQueryLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: GraphPersistenceLatency, Aggregation: view.Distribution(bounds...)},
		{Measure: ProviderNotificationCount, Aggregation: view.Count(), TagKeys: []tag.Key{providerTagKey}},
		{Measure: ProviderPayloadCount, Aggregation: view.Count(), TagKeys: []tag.Key{providerTagKey}},
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

// Counter t=>[tag], c=>[count variable]
func Counter(ctx context.Context, m *stats.Int64Measure, t string, c int64) func() {
	return func() {
		_ = stats.RecordWithOptions(
			ctx,
			stats.WithTags(tag.Insert(providerTagKey, t)),
			stats.WithMeasurements(m.M(c)),
		)
	}
}

var logger = log.NewSubsystemLogger()

// Handler creates an HTTP router for serving metric info
func Handler(views []*view.View) http.Handler {
	// Register default views
	err := view.Register(builtinViews...)
	if err != nil {
		logger.Errorf("cannot register metrics default views: %s", err)
	}
	// Register other views
	err = view.Register(views...)
	if err != nil {
		logger.Errorf("cannot register metrics views: %s", err)
	}
	registry, ok := promclient.DefaultRegisterer.(*promclient.Registry)
	if !ok {
		logger.Warnf("failed to export default prometheus registry; "+
			"some metrics will be unavailable; unexpected types: %T", promclient.DefaultRegisterer)
	}
	exporter, err := prometheus.NewExporter(prometheus.Options{
		Registry:  registry,
		Namespace: "pando",
	})
	if err != nil {
		logger.Errorf("could not create the prometheus stats exporter: %v", err)
	}

	return exporter
}
