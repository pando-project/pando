package legs_interface

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
)

type PandoCore interface {
	Subscribe(ctx context.Context, peerID peer.ID) error
	Unsubscribe(ctx context.Context, peerID peer.ID) error
}
