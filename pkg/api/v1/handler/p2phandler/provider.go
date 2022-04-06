package p2phandler

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	pb "github.com/kenlabs/pando/pkg/api/v1/libp2p/proto"
	"github.com/kenlabs/pando/pkg/api/v1/model"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
)

func (h *libp2pHandler) providerRegister(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	err := h.handler.ProviderRegister(ctx, msg.GetData())
	if err != nil {
		logger.Errorf("register provider failed: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (h *libp2pHandler) listProviderInfo(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	var peerid peer.ID
	err := json.Unmarshal(msg.GetData(), &peerid)
	if err != nil {
		return nil, v1.NewError(fmt.Errorf("failed to unmarshal peerid: %v", err), http.StatusBadRequest)
	}
	info, err := h.handler.ListProviderInfo(peerid)
	if err != nil {
		return nil, err
	}

	resBytes, err := model.GetProviderRes(info)
	if err != nil {
		logger.Errorf("failed to marshal provider info, err: %v", err)
		return nil, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}

	return resBytes, nil
}

func (h *libp2pHandler) listProviderHead(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	var peerid peer.ID
	err := json.Unmarshal(msg.GetData(), &peerid)
	if err != nil {
		return nil, v1.NewError(fmt.Errorf("failed to unmarshal peerid: %v", err), http.StatusBadRequest)
	}
	headCid, err := h.handler.ListProviderHead(peerid)
	if err != nil {
		return nil, err
	}
	return headCid.Bytes(), nil
}
