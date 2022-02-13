package provider

import (
	"github.com/ipfs/go-cid"
)

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	Push(metadataCid cid.Cid) error
}
