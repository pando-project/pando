package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ipfs/go-cid"
	storeError "github.com/kenlabs/pando-store/pkg/error"
	"github.com/kenlabs/pando-store/pkg/types/cbortypes"
	v1 "github.com/pando-project/pando/pkg/api/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/http"
	"strconv"
)

func (c *Controller) SnapShotList() ([]byte, error) {
	res, err := c.Core.StoreInstance.PandoStore.SnapShotStore().GetSnapShotList(context.Background())
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

func (c *Controller) MetadataSnapShot(ctx context.Context, cidstr string, height string) ([]byte, error) {
	var snapshotFromHeight *cbortypes.SnapShot
	var snapshotFromCid *cbortypes.SnapShot
	if cidstr == "" && height == "" {
		return nil, v1.NewError(errors.New("height or cid is required"), http.StatusBadRequest)
	}

	if cidstr != "" {
		snapshotCid, err := cid.Decode(cidstr)
		if err != nil {
			return nil, err
		}
		snapshotFromCid, err = c.Core.StoreInstance.PandoStore.SnapShotStore().GetSnapShotByCid(ctx, snapshotCid)
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
		snapshotFromHeight, _, err = c.Core.StoreInstance.PandoStore.SnapShotStore().GetSnapShotByHeight(ctx, snapshotHeight)
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

func (c *Controller) MetaInclusion(ctx context.Context, cidstr string) ([]byte, error) {
	metaCid, err := cid.Decode(cidstr)
	if err != nil {
		logger.Errorf("invalid cid: %s, err:%v", metaCid.String(), err)
		return nil, v1.NewError(errors.New("invalid cid"), http.StatusBadRequest)
	}

	inclusion, err := c.Core.StoreInstance.PandoStore.MetaInclusion(ctx, metaCid)
	if err != nil {
		logger.Errorf("failed to get meta inclusion for cid: %s, err:%v", metaCid.String(), err)
		return nil, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}

	res, err := json.Marshal(inclusion)
	if err != nil {
		return nil, v1.NewError(err, http.StatusInternalServerError)
	}

	return res, nil
}

func (c *Controller) MetadataQuery(ctx context.Context, providerID string, queryStr string) (queryResult interface{}, err error) {
	var bsonQuery bson.D
	err = bson.UnmarshalExtJSON([]byte(queryStr), true, &bsonQuery)
	if err != nil {
		return nil, err
	}
	var resJson []bson.M
	opts := options.RunCmd().SetReadPreference(readpref.Primary())
	res, err := c.Core.StoreInstance.MetadataCache.Database(providerID).RunCommandCursor(
		ctx,
		bsonQuery,
		opts,
	)
	if err != nil {
		return nil, err
	}
	if err = res.All(ctx, &resJson); err != nil {
		return nil, err
	}

	return resJson, nil
}
