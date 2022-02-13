package http

import (
	"context"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"net/http"
	"pando/pkg/api/types"
	"pando/pkg/api/v1"
	"pando/pkg/metrics"
	"strings"
)

func (a *API) registerPando() {
	pando := a.router.Group("/pando")
	{
		//pando.GET("/subscribe", a.pandoSubscribe)
		pando.GET("/info", a.pandoInfo)
		pando.GET("/metrics", adapter.Wrap(func(h http.Handler) http.Handler {
			return metrics.Handler(coremetrics.DefaultViews)
		}))
		pando.OPTIONS("/health")
	}
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

	ipReplacer := func(multiAddress string, replaceIP string) string {
		splitAddress := strings.Split(multiAddress, "/")
		splitAddress[2] = replaceIP
		return strings.Join(splitAddress, "/")
	}

	apiAddresses := map[string]string{
		"HTTP_API":      ipReplacer(a.options.ServerAddress.HttpAPIListenAddress, a.options.ServerAddress.ExternalIP),
		"GRAPHQL_API":   ipReplacer(a.options.ServerAddress.GraphqlListenAddress, a.options.ServerAddress.ExternalIP),
		"GRAPHSYNC_API": ipReplacer(a.options.ServerAddress.P2PAddress, a.options.ServerAddress.ExternalIP),
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("ok", struct {
		PeerID       string
		APIAddresses map[string]string
	}{pandoInfo.PeerID, apiAddresses}))
}
