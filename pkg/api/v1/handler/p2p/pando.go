package p2p

import (
	"context"
	"encoding/json"
	"github.com/libp2p/go-libp2p-core/peer"
	v1 "github.com/pando-project/pando/pkg/api/v1"
	pb "github.com/pando-project/pando/pkg/api/v1/server/libp2p/proto"
	"net/http"
)

func (h *libp2pHandler) pandoInfo(ctx context.Context, p peer.ID, msg *pb.PandoMessage) ([]byte, error) {
	pandoInfo, err := h.controller.PandoInfo()
	if err != nil {
		return nil, err
	}
	res, err := json.Marshal(pandoInfo)
	if err != nil {
		logger.Errorf("failed to marshal pando info res, err: %v", err)
		return nil, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}
	return res, nil
}
