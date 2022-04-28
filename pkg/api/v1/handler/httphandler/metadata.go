package httphandler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	"github.com/kenlabs/pando/pkg/metrics"
	"net/http"
)

func (a *API) registerMetadata() {
	metadata := a.router.Group("/metadata")
	{
		metadata.GET("/list", a.metadataList)
		metadata.GET("/snapshot", a.metadataSnapshot)
		metadata.GET("/inclusion", a.metadataInclusion)
	}
}

func (a *API) metadataList(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
	defer record()

	snapCidList, err := a.handler.MetadataList()
	if err != nil {
		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", snapCidList))
}

func (a *API) metadataSnapshot(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataSnapshotLatency)
	defer record()

	heightQuery := ctx.Query("height")
	snapshotCidQuery := ctx.Query("cid")

	snapshot, err := a.handler.MetadataSnapShot(ctx, snapshotCidQuery, heightQuery)
	if err != nil {
		logger.Errorf("metadataList metadataSnapshot failed: %v", err)
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))

}

func (a *API) metadataInclusion(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataInclusionLatency)
	defer record()

	cidQuery := ctx.Query("cid")
	inclusion, err := a.handler.MetaInclusion(ctx, cidQuery)
	if err != nil {
		logger.Errorf("failed to get meta inclusion for cid:%s, \nerr:%v", cidQuery, err)
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", inclusion))
}
