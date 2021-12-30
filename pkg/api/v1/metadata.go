package v1

import (
	"context"
	"fmt"
	"net/http"
	"pando/pkg/statetree"
	snapshotTypes "pando/pkg/statetree/types"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"

	"pando/pkg/api/types"
	"pando/pkg/metrics"
)

func (a *API) registerMetadata() {
	metadata := a.router.Group("/metadata")
	{
		metadata.GET("/metadataList", a.metadataList)
		metadata.GET("/metadataSnapshot", a.metadataSnapshot)
	}
}

func (a *API) metadataList(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
	defer record()

	snapCidList, err := a.core.StateTree.GetSnapShotCidList()
	if err != nil {
		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
		handleError(ctx, http.StatusInternalServerError, InternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", snapCidList))
}

func (a *API) metadataSnapshot(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataSnapshotLatency)
	defer record()

	heightQuery := ctx.Query("height")
	snapshotCidQuery := ctx.Query("snapshotCid")

	queryIsNull := true

	var snapshot *snapshotTypes.SnapShot
	var err error

	if snapshotCidQuery != "" {
		queryIsNull = false
		snapshot, err = a.getSnapshotByCid(snapshotCidQuery)
		if err != nil {
			if err == statetree.NotFoundErr {
				handleError(ctx, http.StatusNotFound,
					fmt.Sprintf("metadataSnapshot not found by cid: %s", snapshotCidQuery))
				return
			}

			logger.Errorf("get metadataSnapshot failed: %v", err)
			handleError(ctx, http.StatusInternalServerError, InternalServerError)
			return
		}
	}

	if heightQuery != "" {
		queryIsNull = false
		snapshot, err = a.getSnapshotByHeight(heightQuery)
		if err != nil {
			if err == statetree.NotFoundErr {
				handleError(ctx, http.StatusNotFound,
					fmt.Sprintf("metadataSnapshot not found by height: %s", heightQuery))
				return
			}

			logger.Errorf("get metadataSnapshot failed: %v", err)
			handleError(ctx, http.StatusInternalServerError, InternalServerError)
			return
		}
	}

	if queryIsNull {
		handleError(ctx, http.StatusBadRequest, "height or snapshotCid is required")
		return
	}

	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))

}

func (a *API) getSnapshotByCid(cidStr string) (*snapshotTypes.SnapShot, error) {
	snapshotCid, err := cid.Decode(cidStr)
	if err != nil {
		return nil, err
	}

	snapshot, err := a.core.StateTree.GetSnapShot(snapshotCid)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (a *API) getSnapshotByHeight(heightStr string) (*snapshotTypes.SnapShot, error) {
	snapshotHeight, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return nil, err
	}

	snapshot, err := a.core.StateTree.GetSnapShotByHeight(snapshotHeight)
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}
