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
		metadata.GET("/list", a.metadataList)
		metadata.GET("/snapshot", a.metadataSnapshot)
		metadata.POST("/query", a.metadataQuery)
	}
}

func (a *API) metadataList(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
	defer record()

	snapCidList, err := a.controller.MetadataList()
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

	snapshot, err := a.controller.MetadataSnapShot(ctx, snapshotCidQuery, heightQuery)
	if err != nil {
		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
		HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))

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
