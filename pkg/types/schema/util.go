package schema

import (
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multicodec"
)

// LinkContextKey used to propagate link info through the linkSystem context
type LinkContextKey string

// LinkContextValue used to propagate link info through the linkSystem context
type LinkContextValue bool

const (
	IsMetadataKey = LinkContextKey("isMetadataLink")
)

var LinkProto = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.DagJson),
		MhType:   uint64(multicodec.Sha2_256),
		MhLength: 16,
	},
}

func NewMetadata(payload []byte) Metadata {
	return &_Metadata{
		PreviousID: _Link_Metadata__Maybe{
			m: schema.Maybe_Absent,
		},
		Payload: _Bytes{x: payload},
	}
}

func MetadataLink(lsys ipld.LinkSystem, metadata Metadata) (Link_Metadata, error) {
	lnk, err := lsys.Store(metadata.LinkContext(context.Background()), LinkProto, metadata.Representation())
	if err != nil {
		return nil, err
	}

	return &_Link_Metadata{lnk}, nil
}

func (m Metadata) AppendMetadata(previousID cid.Cid, payload []byte) (Metadata, error) {
	if previousID == cid.Undef {
		return nil, fmt.Errorf("cid is undefined")
	}

	metadataLinkBuilder := Type.Link_Metadata.NewBuilder()
	err := metadataLinkBuilder.AssignLink(cidlink.Link{Cid: previousID})
	if err != nil {
		return nil, fmt.Errorf("failed to generate link from latest metadata cid: %s", err)
	}
	previousLink := metadataLinkBuilder.Build().(Link_Metadata)

	return &_Metadata{
		PreviousID: _Link_Metadata__Maybe{
			m: schema.Maybe_Value,
			v: *previousLink,
		},
		Payload: _Bytes{x: payload},
	}, nil
}

func (m Metadata) LinkContext(ctx context.Context) ipld.LinkContext {
	return ipld.LinkContext{
		Ctx: context.WithValue(ctx, IsMetadataKey, LinkContextValue(true)),
	}
}

func (m Link_Metadata) ToCid() cid.Cid {
	return m.x.(cidlink.Link).Cid
}
