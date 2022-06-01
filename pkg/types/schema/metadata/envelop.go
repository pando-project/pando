package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/record"
)

const (
	metaSignatureCodec  = "/Pando/metaSignature"
	metaSignatureDomain = "Pando"
)

type metaSignatureRecord struct {
	domain *string
	codec  []byte
	metaID []byte
}

func (r *metaSignatureRecord) Domain() string {
	if r.domain != nil {
		return *r.domain
	}
	return metaSignatureDomain
}

func (r *metaSignatureRecord) Codec() []byte {
	if r.codec != nil {
		return r.codec
	}
	return []byte(metaSignatureCodec)
}

func (r *metaSignatureRecord) MarshalRecord() ([]byte, error) {
	return r.metaID, nil
}

func (r *metaSignatureRecord) UnmarshalRecord(buf []byte) error {
	r.metaID = buf
	return nil
}

// VerifyMetadata verifies that the metadata has been signed and
// generated correctly.  Returns the peer ID of the signer.
func VerifyMetadata(meta *Metadata) (peer.ID, error) {
	sig := meta.Signature

	// Consume envelope
	rec := &metaSignatureRecord{}
	envelope, err := record.ConsumeTypedEnvelope(sig, rec)
	if err != nil {
		return peer.ID(""), err
	}

	genID, err := signMetadata(meta)
	if err != nil {
		return peer.ID(""), err
	}

	if !bytes.Equal(genID, rec.metaID) {
		return peer.ID(""), errors.New("invalid signature")
	}

	signerID, err := peer.IDFromPublicKey(envelope.PublicKey)
	if err != nil {
		return peer.ID(""), fmt.Errorf("cannot convert public key to peer ID: %s", err)
	}

	return signerID, nil
}
