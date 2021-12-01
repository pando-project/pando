package http

import (
	"Pando/internal/metrics"
	"Pando/legs"
	"context"
	"fmt"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/gorilla/mux"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"time"

	//"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	//"Pando/internal/httpserver"
	"net/http"
)

// handler handles requests for the finder resource
type httpHandler struct {
	graphSyncHandler *GraphSyncHandler
}

type GraphSyncHandler struct {
	Core *legs.Core
}

func newHandler(core *legs.Core) *httpHandler {
	return &httpHandler{
		graphSyncHandler: &GraphSyncHandler{core},
	}
}

func (h *httpHandler) SubProvider(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	peerID, err := getProviderID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.graphSyncHandler.Core.Subscribe(context.Background(), peerID)
	if err != nil {
		log.Error("cannot create subscriber", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	_ = stats.RecordWithOptions(context.Background(),
		stats.WithTags(tag.Insert(metrics.Method, "api")),
		stats.WithMeasurements(metrics.SubProviderLatency.M(coremetrics.MsecSince(startTime))))
	w.WriteHeader(http.StatusOK)
}

func getProviderID(r *http.Request) (peer.ID, error) {
	vars := mux.Vars(r)
	pid := vars["peerid"]
	providerID, err := peer.Decode(pid)
	if err != nil {
		return providerID, fmt.Errorf("cannot decode provider id: %s", err)
	}
	return providerID, nil
}
