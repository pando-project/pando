package handler

import (
	"encoding/json"
	"errors"
	"github.com/ipfs/go-cid"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/statetree"
	snapshotTypes "github.com/kenlabs/pando/pkg/statetree/types"
	"net/http"
	"strconv"
)

func (h *ServerHandler) MetadataList() ([]byte, error) {
	res, err := h.Core.StateTree.GetSnapShotCidList()
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}
	if res == nil {
		return nil, v1.NewError(v1.ResourceNotFound, http.StatusNotFound)
	}
	data, err := json.Marshal(res)
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}

	return data, nil
}

func (h *ServerHandler) MetadataSnapShot(c string, height string) ([]byte, error) {
	var snapshotFromHeight *snapshotTypes.SnapShot
	var snapshotFromCid *snapshotTypes.SnapShot
	if c == "" && height == "" {
		return nil, v1.NewError(errors.New("height or cid is required"), http.StatusBadRequest)
	}

	if c != "" {
		snapshotCid, err := cid.Decode(c)
		if err != nil {
			return nil, err
		}
		snapshotFromCid, err = h.Core.StateTree.GetSnapShot(snapshotCid)
		if err != nil {
			if err == statetree.NotFoundErr {
				return nil, v1.NewError(v1.ResourceNotFound, http.StatusNotFound)
			}
			return nil, v1.NewError(err, http.StatusInternalServerError)
		}
	}
	if height != "" {
		snapshotHeight, err := strconv.ParseUint(height, 10, 64)
		if err != nil {
			return nil, v1.NewError(err, http.StatusBadRequest)
		}
		snapshotFromHeight, err = h.Core.StateTree.GetSnapShotByHeight(snapshotHeight)
		if err != nil {
			if err == statetree.NotFoundErr {
				return nil, v1.NewError(v1.ResourceNotFound, http.StatusNotFound)
			}
			return nil, v1.NewError(err, http.StatusInternalServerError)
		}
	}
	if snapshotFromHeight != nil && snapshotFromCid != nil && snapshotFromHeight != snapshotFromCid {
		return nil, v1.NewError(errors.New("dismatched cid and height for snapshot"), http.StatusBadRequest)
	}

	var resSnapshot *snapshotTypes.SnapShot
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
