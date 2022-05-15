package pando

import (
	"context"
	coremetrics "github.com/filecoin-project/go-indexer-core/metrics"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/kenlabs/pando/pkg/api/types"
	"github.com/kenlabs/pando/pkg/metrics"
	"net/http"
)

func (a *API) registerPando() {
	pando := a.router.Group("/pando")
	{
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

	pandoInfo, err := a.controller.PandoInfo()
	if err != nil {
		logger.Errorf("get pando pandoInfo failed: %v", err)
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", *pandoInfo))
}
