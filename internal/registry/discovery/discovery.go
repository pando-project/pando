package discovery

import (
	"context"
	"github.com/filecoin-project/lotus/chain/types"

	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// Provider types
	OtherType = iota
	MinerType
)

const (
	Little = "little"
	Normal = "normal"
	Large  = "large"
)

var (
	//LittleAccount = types.NewInt(1)
	NormalAccount = types.NewInt(10)
	LargeAccount  = types.NewInt(100)
)

// Discoverer is the interface that supplies functionality to discover providers
type Discoverer interface {
	Discover(ctx context.Context, peerID peer.ID, discoveryAddr string) (*Discovered, error)
}

// Discovered holds information about a provider that is discovered
type Discovered struct {
	AddrInfo    peer.AddrInfo
	BalanceType string
	Type        int
}
