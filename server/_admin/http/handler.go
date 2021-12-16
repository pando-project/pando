package httpadminserver

import (
	"Pando/internal/metrics"
	"Pando/legs"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
)

// handlers handles requests for the finder resource
type httpHandler struct {
	Core *legs.Core
}

func newHandler(core *legs.Core) *httpHandler {
	return &httpHandler{core}
}

func (h *httpHandler) SubProvider(c *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.SubProviderLatency)
	defer record()

	peerID := c.Query("peerid")
	providerID, err := peer.Decode(peerID)
	if err != nil {
		log.Warnf("cannot decode provider id: %s", err)
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = h.Core.Subscribe(context.Background(), providerID)
	if err != nil {
		log.Error("cannot create subscriber", "err", err)
		http.Error(c.Writer, "", http.StatusInternalServerError)
		return
	}

	c.Writer.WriteHeader(http.StatusOK)
}
