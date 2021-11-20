package account

import (
	"Pando/internal/registry"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PeerType int

const (
	_ PeerType = iota
	UnregisteredPeer
	WhiteListPeer
	RegisteredPeer
)

type Info struct {
	PeerType     PeerType
	AccountLevel int
}

func FetchPeerType(peerID peer.ID, registry *registry.Registry) *Info {
	peerType := UnregisteredPeer
	peerAccountLevel, err := registry.ProviderAccountLevel(peerID)
	if err != nil {
		peerType = RegisteredPeer
	}
	if registry.IsTrusted(peerID) {
		peerType = WhiteListPeer
	}
	return &Info{
		PeerType:     peerType,
		AccountLevel: peerAccountLevel,
	}
}
