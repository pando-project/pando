package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	v1 "github.com/kenlabs/pando/pkg/api/v1"
	"github.com/kenlabs/pando/pkg/api/v1/model"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"net/http"
)

func (c *Controller) ProviderRegister(ctx context.Context, data []byte) error {
	registerRequest, err := model.ReadRegisterRequest(data)
	if err != nil {
		logger.Errorf("read register info failed: %v\n", err)
		return v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}

	if len(registerRequest.PeerID) == 0 {
		logger.Errorf("missing account id")
		return v1.NewError(errors.New("missing account id"), http.StatusBadRequest)
	}

	if err = c.Core.Registry.CheckSequence(registerRequest.PeerID, registerRequest.Seq); err != nil {
		logger.Errorf("bad sequence: %v", err.Error())
		return v1.NewError(fmt.Errorf("bad sequence: %v", err.Error()), http.StatusBadRequest)
	}

	providerMultiAddr := make([]multiaddr.Multiaddr, len(registerRequest.Addrs))
	for i, m := range registerRequest.Addrs {
		var err error
		providerMultiAddr[i], err = multiaddr.NewMultiaddr(m)
		if err != nil {
			logger.Errorf("invalid address: %s", providerMultiAddr[i])
			return v1.NewError(fmt.Errorf("invalid address: %s", providerMultiAddr[i]), http.StatusBadRequest)
		}
	}

	info := &registry.ProviderInfo{
		AddrInfo: peer.AddrInfo{
			ID:    registerRequest.PeerID,
			Addrs: providerMultiAddr,
		},
	}
	err = c.Core.Registry.Register(ctx, info)

	logger.Debugf("pando register success: %s", info.AddrInfo.ID)

	return nil

}

func (c *Controller) ListProviderInfo(p peer.ID) ([]*registry.ProviderInfo, error) {
	info := c.Core.Registry.ProviderInfo(p)
	if info == nil {
		return nil, v1.NewError(errors.New("provider not found"), http.StatusNotFound)
	}
	return info, nil
}

func (c *Controller) ListProviderHead(p peer.ID) (cid.Cid, error) {
	var cidBytes []byte
	var err error
	cidBytes, err = h.Core.StoreInstance.MutexDataStore.Get(context.Background(), datastore.NewKey(legs.SyncPrefix+p.String()))
	if err != nil {
		if err == datastore.ErrNotFound {
			return cid.Undef, v1.NewError(errors.New("not found the head of this provider"), http.StatusNotFound)
		}
		return cid.Undef, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}
	_, providerCid, err := cid.CidFromBytes(cidBytes)
	if err != nil {
		return cid.Undef, v1.NewError(v1.InternalServerError, http.StatusInternalServerError)
	}
	return providerCid, nil
}
