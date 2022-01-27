//go:generate go run gen.go .

package schema

import (
	"context"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
)

// NoEntries is a special value used to explicitly indicate that an
// advertisement does not have any entries. When isRm is true it and serves to
// remove content by context ID, and when isRm is false it serves to update
// metadata only.
var NoEntries cidlink.Link

// Linkproto is the ipld.LinkProtocol used for the ingestion protocol.
// Refer to it if you have encoding questions.
var Linkproto = cidlink.LinkPrototype{
	Prefix: cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.DagJson),
		MhType:   uint64(multicodec.Sha2_256),
		MhLength: 16,
	},
}

var mhCode = multihash.Names["sha2-256"]

func init() {
	// Define NoEntries as the CID of a sha256 hash of nil.
	m, err := multihash.Sum(nil, multihash.SHA2_256, 16)
	if err != nil {
		panic(err)
	}
	NoEntries = cidlink.Link{Cid: cid.NewCidV1(cid.Raw, m)}
}

// LinkContextKey used to propagate link info through the linkSystem context
type LinkContextKey string

// LinkContextValue used to propagate link info through the linkSystem context
type LinkContextValue bool

const (
	// IsMetadataKey is a LinkContextValue that determines the schema type the
	// link belongs to. This is used to understand what callback to trigger
	// in the linksystem when we come across a specific linkType.
	IsMetadataKey = LinkContextKey("isMetadataLink")
	// ContextID must not exceed this number of bytes.
	MaxContextIDLen = 64
)

func mhsToBytes(mhs []multihash.Multihash) []_Bytes {
	out := make([]_Bytes, len(mhs))
	for i := range mhs {
		out[i] = _Bytes{x: mhs[i]}
	}
	return out
}

// LinkAdvFromCid creates a link advertisement from a CID
func LinkAdvFromCid(c cid.Cid) Link_Advertisement {
	return &_Link_Advertisement{x: cidlink.Link{Cid: c}}
}

// ToCid converts a link to CID
func (l Link_Metadata) ToCid() cid.Cid {
	return l.x.(cidlink.Link).Cid
}

// LinkContext returns a linkContext for the type of link
func (l Metadata) LinkContext(ctx context.Context) ipld.LinkContext {
	return ipld.LinkContext{
		Ctx: context.WithValue(ctx, IsMetadataKey, LinkContextValue(true)),
	}
}

// NewListOfMhs is a convenient method to create a new list of bytes
// from a list of multihashes that may be consumed by a linksystem.
func NewListOfMhs(lsys ipld.LinkSystem, mhs []multihash.Multihash) (ipld.Link, error) {
	cStr := &_List_Bytes{x: mhsToBytes(mhs)}
	return lsys.Store(ipld.LinkContext{}, Linkproto, cStr)
}

// NewListBytesFromMhs converts multihashes to a list of bytes
func NewListBytesFromMhs(mhs []multihash.Multihash) List_Bytes {
	return &_List_Bytes{x: mhsToBytes(mhs)}
}

// NewLinkedListOfMhs creates a new element of a linked list that
// can be used to paginate large lists.
func NewLinkedListOfMhs(lsys ipld.LinkSystem, mhs []multihash.Multihash, next ipld.Link) (ipld.Link, EntryChunk, error) {
	cStr := &_EntryChunk{
		Entries: _List_Bytes{x: mhsToBytes(mhs)},
	}
	// If no next in the list.
	if next == nil {
		cStr.Next = _Link_EntryChunk__Maybe{m: schema.Maybe_Absent}
	} else {
		cStr.Next = _Link_EntryChunk__Maybe{m: schema.Maybe_Value, v: _Link_EntryChunk{x: next}}
	}

	lnk, err := lsys.Store(ipld.LinkContext{}, Linkproto, cStr.Representation())
	return lnk, cStr, err
}

func NewMetadata(previousID Link_Metadata) (Metadata, error) {
	metadata := &_Metadata{}
	if previousID == nil {
		metadata.PreviousID = _Link_Metadata__Maybe{m: schema.Maybe_Absent}
	} else {
		metadata.PreviousID = _Link_Metadata__Maybe{m: schema.Maybe_Value, v: *previousID}
	}

	return metadata, nil
}

func MetadataLink(lsys ipld.LinkSystem, metadata Metadata) (Link_Metadata, error) {
	lnk, err := lsys.Store(metadata.LinkContext(context.Background()), Linkproto, metadata.Representation())
	if err != nil {
		return nil, err
	}

	return &_Link_Metadata{lnk}, nil
}
