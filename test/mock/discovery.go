package mock

import (
	"Pando/config"
	"Pando/internal/registry/discovery"
	"context"
	"errors"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"time"
)

type mockDiscoverer struct {
	discoverRsp *discovery.Discovered
}

const (
	exceptID   = "12D3KooWK7CTS7cyWi51PeNE3cTjS2F2kDCZaQVU4A5xBmb9J1do"
	trustedID  = "12D3KooWSG3JuvEjRkSxt93ADTjQxqe4ExbBwSkQ9Zyk1WfBaZJF"
	trustedID2 = "12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr"

	minerDiscoAddr = "stitest999999"
	minerAddr      = "/ip4/127.0.0.1/tcp/9999"
	minerAddr2     = "/ip4/127.0.0.2/tcp/9999"
)

var discoveryCfg = config.Discovery{
	Policy: config.Policy{
		Allow:       false,
		Except:      []string{exceptID, trustedID, trustedID2},
		Trust:       false,
		TrustExcept: []string{trustedID, trustedID2},
	},
	PollInterval:   config.Duration(time.Minute),
	RediscoverWait: config.Duration(time.Minute),
}

var aclCfg = config.AccountLevel{Threshold: []int{1, 10, 99}}

func newMockDiscoverer(providerID string) (*mockDiscoverer, error) {
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
			Type: discovery.MinerType,
		},
	}, nil
}

func (m *mockDiscoverer) Discover(ctx context.Context, peerID peer.ID, filecoinAddr string) (*discovery.Discovered, error) {
	if filecoinAddr == "bad1234" {
		return nil, errors.New("unknown miner")
	}

	return m.discoverRsp, nil
}
