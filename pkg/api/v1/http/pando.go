package http

import (
	"context"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
	"pando/pkg/api/types"
	"pando/pkg/api/v1"
	"pando/pkg/metrics"
	"strings"
)

func (a *API) registerPando() {
	pando := a.router.Group("/pando")
	{
		pando.GET("/subscribe", a.pandoSubscribe)
		pando.GET("/info", a.pandoInfo)
		pando.GET("/metrics", adapter.Wrap(func(h http.Handler) http.Handler {
			return metrics.Handler(coremetrics.DefaultViews)
		}))
		pando.OPTIONS("/health")
	}
}

func (a *API) pandoSubscribe(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetPandoSubscribeLatency)
	defer record()

	providerQuery := ctx.Query("provider")
	providerPeerID, err := peer.Decode(providerQuery)
	if err != nil {
		logger.Errorf("decode provider peerID failed: %v\n", err)
		handleError(ctx, http.StatusBadRequest, "peerID of provider is invalid")
		return
	}

	err = a.core.LegsCore.Subscribe(context.Background(), providerPeerID)
	if err != nil {
		logger.Errorf("subscribe provider failed: %v\n", err)
		handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("subscribe success", nil))
}

func (a *API) pandoInfo(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetPandoInfoLatency)
	defer record()

	pandoInfo, err := a.core.StateTree.GetPandoInfo()
	if err != nil {
		logger.Errorf("get pando pandoInfo failed: %v", err)
		handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
		return
	}

	multiAddresses := strings.Fields(pandoInfo.MultiAddresses)
	ctx.JSON(http.StatusOK, types.NewOKResponse("ok", struct {
		PeerID         string
		MultiAddresses []string
	}{pandoInfo.PeerID, multiAddresses}))
}
