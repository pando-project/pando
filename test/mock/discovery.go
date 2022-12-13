package mock

import (
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/registry"
	"github.com/pando-project/pando/pkg/registry/discovery"
	"math/big"
	"time"
)

type mockDiscoverer struct {
	discoverRsp *discovery.Discovered
}

const (
	exceptID  = "12D3KooWK7CTS7cyWi51PeNE3cTjS2F2kDCZaQVU4A5xBmb9J1do"
	blackID   = "12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr"
	minerAddr = "/ip4/127.0.0.1/tcp/9999"
)

var MockDiscoveryCfg = option.Discovery{
	Policy: option.Policy{
		Allow:       true,
		Except:      []string{blackID},
		Trust:       true,
		TrustExcept: []string{},
	},
	PollInterval:   option.Duration(time.Second * 2).String(),
	RediscoverWait: option.Duration(time.Minute).String(),
}

var MockAclCfg = option.AccountLevel{Threshold: []int{1, 10, 99}}

func NewMockDiscoverer(providerID string) (*mockDiscoverer, error) {
	peerID, err := peer.Decode(providerID)
	if err != nil {
		return nil, err
	}

	maddr, err := multiaddr.NewMultiaddr(minerAddr)
	if err != nil {
		return nil, err
	}

	return &mockDiscoverer{
		discoverRsp: &discovery.Discovered{
			AddrInfo: peer.AddrInfo{
				ID:    peerID,
				Addrs: []multiaddr.Multiaddr{maddr},
			},
			Type:    discovery.MinerType,
			Balance: big.NewInt(0).Mul(registry.FIL, big.NewInt(5)),
		},
	}, nil
}

func (m *mockDiscoverer) Discover(ctx context.Context, peerID peer.ID, filecoinAddr string) (*discovery.Discovered, error) {
	if filecoinAddr == "bad1234" {
		return nil, errors.New("unknown miner")
	}

	return m.discoverRsp, nil
}
