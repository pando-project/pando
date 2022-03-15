package admin

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando/pkg/api/types"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
	"os"
	"path"
	"time"
)

func (a *API) registerBackup() {
	admin := a.router.Group("")
	{
		admin.GET("/backup", a.backupMeta)
	}
}

func (a *API) backupMeta(ctx *gin.Context) {
	record := metrics.APITimer(context.Background(), metrics.GetBackupMetaLatency)
	defer record()

	backInfo, err := decodeBackupInfo(ctx)
	if err != nil || backInfo == nil {
		handleError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid start/end cid for backup"))
		return
	} else {
		if err := backInfo.Provider.Validate(); err != nil {
			// if not set provider id, use Pando's id
			backInfo.Provider = a.core.LegsCore.Host.ID()
		}
		var filePath string
		fileName := fmt.Sprintf(metadata.BackFileName, backInfo.Provider, time.Now().UnixNano())

		// the force dir will not be back up by Pando backupSys auto
		if backInfo.isForce {
			filePath = path.Join(metadata.BackupTmpPath, "force", fileName)
			// clean tmp car file
			defer func() {
				err = os.Remove(filePath)
				if err != nil {
					logger.Errorf("failed to clean the car file, err: %v", err)
				}
			}()
		} else {
			filePath = path.Join(metadata.BackupTmpPath, fileName)
		}

		err = a.core.MetaManager.ExportMetaCar(ctx, filePath, backInfo.End, backInfo.start)
		if err != nil {
			logger.Errorf("failed to generate car file, err:%v", err)
			handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
			return
		}
		// back up right now. If false, pando will back up in the config time
		if backInfo.isForce {
			estId, err := a.core.MetaManager.EstBackupSys.BackupToEstuary(filePath)
			if err != nil {
				logger.Errorf("failed to back up car to estuary, err:%v", err)
				handleError(ctx, http.StatusInternalServerError, v1.InternalServerError)
				return
			}
			ctx.JSON(http.StatusOK, types.NewOKResponse(fmt.Sprintf("back up successfully! Estuary id: %d", estId), ""))
			return
		}

		ctx.JSON(http.StatusOK, types.NewOKResponse("back up meta as car file successfully! Please wait for Pando to back up it into Estuary", ""))

	}

}

type backupInfo struct {
	start    cid.Cid
	End      cid.Cid
	Provider peer.ID
	isForce  bool
}

func decodeBackupInfo(ctx *gin.Context) (*backupInfo, error) {
	start := ctx.Query("start")
	end := ctx.Query("end")
	force := ctx.Query("force")
	provider := ctx.Query("provider")
	scid, err := cid.Decode(start)
	if err != nil {
		return nil, err
	}
	ecid, err := cid.Decode(end)
	if err != nil {
		return nil, err
	}

	var providerID peer.ID
	if provider != "" {
		providerID, err = peer.Decode(provider)
		if err != nil {
			return nil, err
		}
	}

	var isForce bool
	if force == "" {
		isForce = false
	} else if force == "1" {
		isForce = true
	}
	return &backupInfo{
		start:    scid,
		End:      ecid,
		Provider: providerID,
		isForce:  isForce,
	}, nil
}
