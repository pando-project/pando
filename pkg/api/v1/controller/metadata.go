package controller

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

func (c *Controller) MetadataList() ([]byte, error) {
	res, err := c.Core.StateTree.GetSnapShotCidList()
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

func (c *Controller) MetadataSnapShot(snapshotCid string, height string) ([]byte, error) {
	var snapshotFromHeight *snapshotTypes.SnapShot
	var snapshotFromCid *snapshotTypes.SnapShot
	if snapshotCid == "" && height == "" {
		return nil, v1.NewError(errors.New("height or cid is required"), http.StatusBadRequest)
	}

	if snapshotCid != "" {
		snapshotCid, err := cid.Decode(snapshotCid)
		if err != nil {
			return nil, err
		}
		snapshotFromCid, err = c.Core.StateTree.GetSnapShot(snapshotCid)
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
		snapshotFromHeight, err = c.Core.StateTree.GetSnapShotByHeight(snapshotHeight)
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
