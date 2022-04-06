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
	}
}

func (a *API) metadataList(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
	defer record()

	snapCidList, err := a.handler.MetadataList()
	if err != nil {
		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", snapCidList))
}

//func (a *API) metadataList(ctx *gin.Context) {
//	record := metrics.APITimer(context.Background(), metrics.GetMetadataListLatency)
//	defer record()
//
//	snapCidList, err := a.core.StateTree.GetSnapShotCidList()
//	if err != nil {
//		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
//		handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
//		return
//	}
//
//	ctx.JSON(http.StatusOK, types.NewOKResponse("OK", snapCidList))
//}

func (a *API) metadataSnapshot(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetMetadataSnapshotLatency)
	defer record()

	heightQuery := ctx.Query("height")
	snapshotCidQuery := ctx.Query("cid")

	snapshot, err := a.handler.MetadataSnapShot(snapshotCidQuery, heightQuery)
	if err != nil {
		logger.Error(fmt.Sprintf("metadataList metadataSnapshot failed: %v", err))
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))

}

//func (a *API) metadataSnapshot(ctx *gin.Context) {
//	record := metrics.APITimer(context.Background(), metrics.GetMetadataSnapshotLatency)
//	defer record()
//
//	heightQuery := ctx.Query("height")
//	snapshotCidQuery := ctx.Query("cid")
//
//	var snapshot *snapshotTypes.SnapShot
//	var err error
//
//	if snapshotCidQuery != "" {
//		snapshot, err = a.getSnapshotByCid(snapshotCidQuery)
//		if err != nil {
//			if err == statetree.NotFoundErr {
//				handleError(ctx, http.StatusNotFound,
//					fmt.Errorf("metadataSnapshot not found by cid: %s", snapshotCidQuery))
//				return
//			}
//
//			logger.Errorf("get metadataSnapshot failed: %v", err)
//			handleError(ctx, http.StatusBadRequest, fmt.Errorf("invalid cid: %v", err))
//			return
//		}
//	} else if heightQuery != "" {
//		snapshot, err = a.getSnapshotByHeight(heightQuery)
//		if err != nil {
//			if err == statetree.NotFoundErr {
//				handleError(ctx, http.StatusNotFound,
//					fmt.Errorf("metadataSnapshot not found by height: %s", heightQuery))
//				return
//			}
//
//			logger.Errorf("get metadataSnapshot failed: %v", err)
//			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
//			return
//		}
//	} else {
//		handleError(ctx, http.StatusBadRequest, errors.New("height or snapshotCid is required"))
//		return
//	}
//
//	ctx.JSON(http.StatusOK, types.NewOKResponse("metadataSnapshot found", snapshot))
//}

//func (a *API) getSnapshotByCid(cidStr string) (*snapshotTypes.SnapShot, error) {
//	snapshotCid, err := cid.Decode(cidStr)
//	if err != nil {
//		return nil, err
//	}
//
//	snapshot, err := a.core.StateTree.GetSnapShot(snapshotCid)
//	if err != nil {
//		return nil, err
//	}
//
//	return snapshot, nil
//}

//func (a *API) getSnapshotByHeight(heightStr string) (*snapshotTypes.SnapShot, error) {
//	snapshotHeight, err := strconv.ParseUint(heightStr, 10, 64)
//	if err != nil {
//		return nil, err
//	}
//
//	snapshot, err := a.core.StateTree.GetSnapShotByHeight(snapshotHeight)
//	if err != nil {
//		return nil, err
//	}
//
//	return snapshot, nil
//}
