package consumer

import (
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
)

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	GetLatestHead(provider string) (cid.Cid, error)
	GetLatestSync() cid.Cid
	Sync(nextCid cid.Cid, selector ipld.Node) (cid.Cid, error)
}
