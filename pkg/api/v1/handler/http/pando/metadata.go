package pando

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kenlabs/pando/pkg/api/types"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/metrics"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"io/ioutil"
	"net/http"
)

func (a *API) registerMetadata() {
	metadata := a.router.Group("/metadata")
	{
		metadata.GET("/list", a.snapShotList)
		metadata.GET("/snapshot", a.metadataSnapshot)
		metadata.GET("/inclusion", a.metaInclusion)
		metadata.POST("/query", a.metadataQuery)
	}
}

func (a *API) snapShotList(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
	defer record()

	snapCidList, err := a.controller.SnapShotList()
	if err != nil {
		logger.Error(fmt.Sprintf("snapShotList metadataSnapshot failed: %v", err))
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

	snapshot, err := a.controller.MetadataSnapShot(ctx, snapshotCidQuery, heightQuery)
	if err != nil {
		logger.Error(fmt.Sprintf("snapShotList metadataSnapshot failed: %v", err))
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))

}

func (a *API) metaInclusion(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataInclusionLatency)
	defer record()

	cidQuery := ctx.Query("cid")

	inclusion, err := a.controller.MetaInclusion(ctx, cidQuery)
	if err != nil {
		logger.Error(fmt.Sprintf("get metaInclusion failed: %v", err))
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("metaInclusion found", inclusion))
}

type MetadataQueryRequestBody struct {
	ProviderID string `json:"ProviderID"`
	BsonQuery  string `json:"BsonQuery"`
}

func (a *API) metadataQuery(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.PostMetadataQueryLatency)
	defer record()

	bodyBytes, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		logger.Errorf("read metadata query body failed: %v\n", err)
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	var queryBody *MetadataQueryRequestBody
	if err = json.Unmarshal(bodyBytes, &queryBody); err != nil {
		if err == bsonrw.ErrInvalidJSON {
			HandleError(ctx, v1.NewError(v1.InvalidQuery, http.StatusBadRequest))
			return
		}
		HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
		return
	}

	results, err := a.controller.MetadataQuery(context.Background(), queryBody.ProviderID, queryBody.BsonQuery)
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", results))
}
