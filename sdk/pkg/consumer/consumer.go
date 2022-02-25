package consumer

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
)

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	GetHead(ctx context.Context) (cid.Cid, error)
	Sync(ctx context.Context, nextCid cid.Cid, selector ipld.Node) error
}
