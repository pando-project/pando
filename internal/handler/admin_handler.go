package handler

import (
	"Pando/api/v0/admin/model"
	"Pando/internal/registry"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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

	maddrs := make([]multiaddr.Multiaddr, len(peerRec.Addrs))
	for i, m := range peerRec.Addrs {
		var err error
		maddrs[i], err = multiaddr.NewMultiaddr(m)
		if err != nil {
			return fmt.Errorf("bad address: %s", err)
		}
	}

	info := &registry.ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    peerRec.PeerID,
			Addrs: maddrs,
		},
		DiscoveryAddr: peerRec.MinerAccount,
	}
	return h.registry.Register(info)
}
