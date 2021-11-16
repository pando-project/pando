package handler

import (
	"Pando/internal/registry"
	"errors"
	"fmt"
	"github.com/filecoin-project/storetheindex/api/v0/ingest/model"
	"github.com/libp2p/go-libp2p-core/peer"
)

type AdminHandler struct {
	registry *registry.Registry
}

func NewAdminHandler(registry *registry.Registry) *AdminHandler {
	return &AdminHandler{registry: registry}
}

func (h *AdminHandler) RegisterProvider(data []byte) error {
	peerRec, err := model.ReadRegisterRequest(data)
	if err != nil {
		return fmt.Errorf("cannot read register request: %s", err)
	}

	if len(peerRec.PeerID) == 0 {
		return errors.New("missing peer id")
	}

	if err = h.registry.CheckSequence(peerRec.PeerID, peerRec.Seq); err != nil {
		return err
	}

	info := &registry.ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerRec.PeerID,
			Addrs: peerRec.Addrs,
		},
	}
	return h.registry.Register(info)
}
