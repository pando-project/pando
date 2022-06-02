package metadata

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/record"
)

// LinkContextKey used to propagate link info through the linkSystem context
type LinkContextKey string

// LinkContextValue used to propagate link info through the linkSystem context
type LinkContextValue bool

const (
	IsMetadataKey = LinkContextKey("isMetadataLink")
)

func NewMetaWithBytesPayload(payload []byte, provider peer.ID, signKey crypto.PrivKey) (*Metadata, error) {

	pnode := basicnode.NewBytes(payload)
	meta := &Metadata{
		PreviousID: nil,
		Provider:   provider.String(),
		Payload:    pnode,
	}

	sig, err := SignWithPrivateKey(signKey, meta)
	if err != nil {
		return nil, err
	}

	// Add signature
	meta.Signature = sig
	return meta, nil
}

func NewMetaWithPayloadNode(payload datamodel.Node, provider peer.ID, signKey crypto.PrivKey, prev datamodel.Link) (*Metadata, error) {
	meta := &Metadata{
		Provider: provider.String(),
		Payload:  payload,
	}
	if prev == nil {
		meta.PreviousID = nil
	} else {
		meta.PreviousID = &prev
	}

	sig, err := SignWithPrivateKey(signKey, meta)
	if err != nil {
		return nil, err
	}

	// Add signature
	meta.Signature = sig
	return meta, nil
}

func NewMetadataWithLink(payload []byte, provider peer.ID, signKey crypto.PrivKey, link datamodel.Link) (*Metadata, error) {
	if link == nil {
		return nil, fmt.Errorf("nil previous meta link")
	}

	pnode := basicnode.NewBytes(payload)
	meta := &Metadata{
		PreviousID: &link,
		Provider:   provider.String(),
		Payload:    pnode,
	}

	sig, err := SignWithPrivateKey(signKey, meta)
	if err != nil {
		return nil, err
	}

	// Add signature
	meta.Signature = sig

	return meta, nil
}

func MetadataLink(lsys ipld.LinkSystem, metadata *Metadata) (datamodel.Link, error) {
	mnode, err := metadata.ToNode()
	if err != nil {
		return cidlink.Link{}, err
	}
	lnk, err := lsys.Store(metadata.LinkContext(context.Background()), LinkProto, mnode)
	if err != nil {
		return nil, err
	}

	return lnk, nil
}

func (m Metadata) AppendMetadata(previousID cid.Cid, provider peer.ID, payload []byte, signKey crypto.PrivKey) (*Metadata, error) {
	if previousID == cid.Undef {
		return nil, fmt.Errorf("cid is undefined")
	}

	pnode := basicnode.NewBytes(payload)
	lk := ipld.Link(cidlink.Link{Cid: previousID})

	metadata := &Metadata{
		PreviousID: &lk,
		Provider:   provider.String(),
		Payload:    pnode,
	}
	signature, err := SignWithPrivateKey(signKey, metadata)
	if err != nil {
		return nil, err
	}
	metadata.Signature = signature

	return metadata, nil
}

func (m Metadata) LinkContext(ctx context.Context) ipld.LinkContext {
	return ipld.LinkContext{
		Ctx: context.WithValue(ctx, IsMetadataKey, LinkContextValue(true)),
	}
}

// SignWithPrivateKey Signs metadata using libp2p envelope
func SignWithPrivateKey(privateKey crypto.PrivKey, meta *Metadata) ([]byte, error) {
	metaID, err := signMetadata(meta)
	if err != nil {
		return nil, err
	}
	envelope, err := record.Seal(&metaSignatureRecord{metaID: metaID}, privateKey)
	if err != nil {
		return nil, err
	}
	return envelope.Marshal()
}

func signMetadata(meta *Metadata) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	m := &Metadata{
		PreviousID: meta.PreviousID,
		Provider:   meta.Provider,
		Payload:    meta.Payload,
	}
	n, err := m.ToNode()
	if err != nil {
		return nil, err
	}

	err = dagjson.Encode(n, buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
