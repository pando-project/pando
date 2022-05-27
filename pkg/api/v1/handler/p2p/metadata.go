package p2phandler

import (
	"context"
	"encoding/json"
	"errors"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	pb "github.com/kenlabs/pando/pkg/api/v1/server/libp2p/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
)

type SnapShotQuery struct {
	CidString string
	Height    string
}

func (h *libp2pHandler) metadataList(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	data, err := h.controller.MetadataList()
	return data, err
}

func (h *libp2pHandler) metadataSnapShot(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	var query *SnapShotQuery
	err := json.Unmarshal(msg.GetData(), query)
	if err != nil {
		logger.Errorw("error unmarshalling metadataSnapShot request", "err", err)
		return nil, v1.NewError(errors.New("cannot decode request"), http.StatusBadRequest)
	}
	data, err := h.controller.MetadataSnapShot(ctx, query.CidString, query.Height)
	if err != nil {
		logger.Errorf("failed to get snapshot: %v", err)
		return nil, err
	}
	return data, nil
}
