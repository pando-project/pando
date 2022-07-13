package provider

import (
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando-store/pkg/types/store"
	"github.com/kenlabs/pando/pkg/types/schema"
)

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	Push(metadata schema.Meta) (cid.Cid, error)
	CheckMetaState(c cid.Cid) (*store.MetaInclusion, error)
}
