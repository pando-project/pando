package provider

import (
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando/pkg/types/schema"
)

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	Push(metadata *schema.Metadata) (cid.Cid, error)
}
