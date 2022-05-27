package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ipfs/go-cid"
	storeError "github.com/kenlabs/PandoStore/pkg/error"
	"github.com/kenlabs/PandoStore/pkg/types/cbortypes"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"net/http"
	"strconv"
)

func (h *ServerHandler) MetadataList() ([]byte, error) {
	res, err := h.Core.StoreInstance.PandoStore.SnapShotStore.GetSnapShotList(context.Background())
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}
	if res == nil {
		return nil, v1.NewError(v1.ResourceNotFound, http.StatusNotFound)
	}
	data, err := json.Marshal(*res)
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}

	return data, nil
}

func (h *ServerHandler) MetadataSnapShot(ctx context.Context, c string, height string) ([]byte, error) {
	var snapshotFromHeight *cbortypes.SnapShot
	var snapshotFromCid *cbortypes.SnapShot
	if c == "" && height == "" {
		return nil, v1.NewError(errors.New("height or cid is required"), http.StatusBadRequest)
	}

	if c != "" {
		snapshotCid, err := cid.Decode(c)
		if err != nil {
			return nil, err
		}
		snapshotFromCid, err = h.Core.StoreInstance.PandoStore.SnapShotStore.GetSnapShotByCid(ctx, snapshotCid)
		if err != nil {
			if err == storeError.InvalidParameters {
				return nil, v1.NewError(v1.InvalidQuery, http.StatusBadRequest)
			}
			return nil, v1.NewError(err, http.StatusInternalServerError)
		}
	}
	if height != "" {
		snapshotHeight, err := strconv.ParseUint(height, 10, 64)
		if err != nil {
			return nil, v1.NewError(err, http.StatusBadRequest)
		}
		snapshotFromHeight, _, err = h.Core.StoreInstance.PandoStore.SnapShotStore.GetSnapShotByHeight(ctx, snapshotHeight)
		if err != nil {
			if err == storeError.InvalidParameters {
				return nil, v1.NewError(v1.InvalidQuery, http.StatusBadRequest)
			}
			return nil, v1.NewError(err, http.StatusInternalServerError)
		}
	}
	if snapshotFromHeight != nil && snapshotFromCid != nil && snapshotFromHeight != snapshotFromCid {
		return nil, v1.NewError(errors.New("dismatched cid and height for snapshot"), http.StatusBadRequest)
	}

	var resSnapshot *cbortypes.SnapShot
	var res []byte
	var err error
	if snapshotFromCid != nil {
		resSnapshot = snapshotFromCid
	} else {
		resSnapshot = snapshotFromHeight
	}

	res, err = json.Marshal(resSnapshot)
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}

	return res, nil
}

func (a *ServerHandler) MetaInclusion(ctx context.Context, metaCid string) ([]byte, error) {
	c, err := cid.Decode(metaCid)
	if err != nil {
		logger.Errorf("invalid cid: %s, err:%v", c.String(), err)
		return nil, v1.NewError(errors.New("invalid cid"), http.StatusBadRequest)
	}

	inclusion, err := a.Core.StoreInstance.PandoStore.MetaInclusion(ctx, c)
	if err != nil {
		logger.Errorf("failed to get meta inclusion for cid: %s, err:%v", c.String(), err)
		return nil, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}

	res, err := json.Marshal(inclusion)
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}

	return res, nil
}
