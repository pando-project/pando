package pkg

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type PandoNode struct {
	PeerAddress string
	PeerID      string
}

func NewPandoPeerInfo(peerAddress string, peerID string) (*peer.AddrInfo, error) {
	multiAddress, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/ipfs/%s", peerAddress, peerID))
	if err != nil {
		return nil, err
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddress)
	if err != nil {
		return nil, err
	}

	return peerInfo, nil
}
