package statetree

import (
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type StateTree struct {
	ProviderState map[peer.ID]*ProviderState
}

type ProviderState struct {
	TotalMetaCount int
	MetaRoot       cid.Cid
}
