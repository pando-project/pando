package admin

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ipfs/go-cid"
	"github.com/pando-project/pando/pkg/api/types"
	"github.com/pando-project/pando/pkg/api/v1"
	"github.com/pando-project/pando/pkg/api/v1/handler/http/pando"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pando-project/pando/pkg/metadata"
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
	backInfo, err := decodeBackupInfo(ctx)
	if err != nil || backInfo == nil {
		pando.HandleError(ctx, v1.NewError(errors.New("invalid start/end cid for backup"), http.StatusBadRequest))
		return
	} else {
		if err := backInfo.provider.Validate(); err != nil {
			// if not set provider id, use Pando's id
			backInfo.provider = a.core.LegsCore.Host.ID()
		}
		var filePath string
		fileName := fmt.Sprintf(metadata.BackFileName, backInfo.provider, time.Now().UnixNano())

		// the force dir will not be back up by Pando backupSys auto
		if backInfo.isForce {
			filePath = path.Join(metadata.BackupTmpPath, "force", fileName)
			// clean tmp car file
			defer func() {
				err = os.Remove(filePath)
				if err != nil {
					logger.Errorf("failed to clean the car file: %s, err: %v", filePath, err)
				}
			}()
		} else {
			filePath = path.Join(metadata.BackupTmpPath, fileName)
		}

		err = a.core.MetaManager.ExportMetaCar(ctx, filePath, backInfo.end, backInfo.start)
		if err != nil {
			logger.Errorf("failed to generate car file start: %s end : %s filepath: %s\r\n, err:%v",
				backInfo.start, backInfo.end, filePath, err)
			pando.HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
			return
		}
		// back up right now. If false, pando will back up in the config time
		if backInfo.isForce {
			estId, err := a.core.MetaManager.EstBackupSys.BackupToEstuary(filePath)
			if err != nil {
				logger.Errorf("failed to back up car file(%s) to estuary, err:%v", filePath, err)
				pando.HandleError(ctx, v1.NewError(v1.InternalServerError, http.StatusInternalServerError))
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
	end      cid.Cid
	provider peer.ID
	isForce  bool
}

func decodeBackupInfo(ctx *gin.Context) (*backupInfo, error) {
	start := ctx.Query("start")
	end := ctx.Query("end")
	force := ctx.Query("force")
	provider := ctx.Query("provider")
	if start == "" || end == "" {
		logger.Errorf("nil start or end cid for backup")
		return nil, fmt.Errorf("nil start or end cid for backup")
	}
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
		end:      ecid,
		provider: providerID,
		isForce:  isForce,
	}, nil
}
