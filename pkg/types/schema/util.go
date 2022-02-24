//go:generate go run gen.go .

package schema

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/schema"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/record"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
)

var mhCode = multihash.Names["sha2-256"]

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

func NewMetadata(payload []byte, provider peer.ID, signKey crypto.PrivKey) (Metadata, error) {

	meta := &_Metadata{
		PreviousID: _Link_Metadata__Maybe{
			m: schema.Maybe_Null,
		},
		Provider: _String{x: provider.String()},
		Payload:  _Bytes{x: payload},
	}

	sig, err := signMetadata(signKey, meta)
	if err != nil {
		return nil, err
	}

	// Add signature
	meta.Signature = _Bytes{x: sig}
	return meta, nil
}

func NewMetadataWithLink(payload []byte, provider peer.ID, signKey crypto.PrivKey, link datamodel.Link) (Metadata, error) {
	if link == nil {
		return nil, fmt.Errorf("nil previous meta link")
	}

	meta := &_Metadata{
		PreviousID: _Link_Metadata__Maybe{
			m: schema.Maybe_Value,
			v: _Link_Metadata{x: link},
		},
		Provider: _String{x: provider.String()},
		Payload:  _Bytes{x: payload},
	}

	sig, err := signMetadata(signKey, meta)
	if err != nil {
		return nil, err
	}

	// Add signature
	meta.Signature = _Bytes{x: sig}

	return meta, nil
}

func MetadataLink(lsys ipld.LinkSystem, metadata Metadata) (Link_Metadata, error) {
	lnk, err := lsys.Store(metadata.LinkContext(context.Background()), LinkProto, metadata.Representation())
	if err != nil {
		return nil, err
	}

	return &_Link_Metadata{lnk}, nil
}

func (m Metadata) AppendMetadata(previousID cid.Cid, provider peer.ID, payload []byte, signKey crypto.PrivKey) (Metadata, error) {
	if previousID == cid.Undef {
		return nil, fmt.Errorf("cid is undefined")
	}

	metadataLinkBuilder := Type.Link_Metadata.NewBuilder()
	err := metadataLinkBuilder.AssignLink(cidlink.Link{Cid: previousID})
	if err != nil {
		return nil, fmt.Errorf("failed to generate link from latest metadata cid: %s", err)
	}
	previousLink := metadataLinkBuilder.Build().(Link_Metadata)

	metadata := &_Metadata{
		PreviousID: _Link_Metadata__Maybe{
			m: schema.Maybe_Value,
			v: *previousLink,
		},
		Provider: _String{x: provider.String()},
		Payload:  _Bytes{x: payload},
	}
	signature, err := signMetadata(signKey, metadata)
	if err != nil {
		return nil, err
	}
	metadata.Signature = _Bytes{x: signature}

	return metadata, nil
}

func (m Metadata) LinkContext(ctx context.Context) ipld.LinkContext {
	return ipld.LinkContext{
		Ctx: context.WithValue(ctx, IsMetadataKey, LinkContextValue(true)),
	}
}

func (m Link_Metadata) ToCid() cid.Cid {
	return m.x.(cidlink.Link).Cid
}

// Signs metadata using libp2p envelope
func signMetadata(privkey crypto.PrivKey, meta Metadata) ([]byte, error) {
	previousID := meta.FieldPreviousID().v

	data := meta.FieldPayload().x
	provider := meta.FieldProvider().x

	advID, err := signatureMetadata(&previousID, provider, data)
	if err != nil {
		return nil, err
	}
	envelope, err := record.Seal(&metaSignatureRecord{advID: advID}, privkey)
	if err != nil {
		return nil, err
	}
	return envelope.Marshal()
}

// Generates the data payload used for signature.
func signatureMetadata(previousID Link_Metadata, prov string, data []byte) ([]byte, error) {
	bindex := cid.Undef.Bytes()
	lindex, err := previousID.AsLink()
	if err != nil {
		return nil, err
	}
	if lindex != nil {
		bindex = lindex.(cidlink.Link).Cid.Bytes()
	}

	// Signature data is previousID+entries+metadata+isRm
	var sigBuf bytes.Buffer
	sigBuf.Grow(len(bindex) + len(data) + len(prov))
	sigBuf.Write(bindex)
	sigBuf.Write(data)
	sigBuf.WriteString(prov)

	return multihash.Sum(sigBuf.Bytes(), mhCode, -1)
}
