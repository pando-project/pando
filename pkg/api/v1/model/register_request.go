package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/record"
)

type RegisterRequest struct {
	// PeerID is the ID of the account this record pertains to.
	PeerID peer.ID

	// Addrs contains the public addresses of the account this record pertains to.
	Addrs []string

	// Seq is a monotonically-increasing sequence counter that's used to order
	// PeerRecords in time. The interval between Seq values is unspecified,
	// but newer PeerRecords MUST have a greater Seq value than older records
	// for the same account.
	Seq uint64

	// filecoin miner account
	MinerAccount string
}

const RegisterRequestEnvelopeDomain = "pando-register-request-record"

var RegisterRequestEnvelopePayloadType = []byte("pando-register-request")

func init() {
	record.RegisterType(&RegisterRequest{})
}

// Domain is used when signing and validating IngestRequest records contained in Envelopes
func (r *RegisterRequest) Domain() string {
	return RegisterRequestEnvelopeDomain
}

// Codec is a binary identifier for the IngestRequest types
func (r *RegisterRequest) Codec() []byte {
	return RegisterRequestEnvelopePayloadType
}

// UnmarshalRecord parses an IngestRequest from a byte slice.
func (r *RegisterRequest) UnmarshalRecord(data []byte) error {
	if r == nil {
		return fmt.Errorf("cannot unmarshal IngestRequest to nil receiver")
	}

	return json.Unmarshal(data, r)
}

// MarshalRecord serializes an IngestRequesr to a byte slice.
func (r *RegisterRequest) MarshalRecord() ([]byte, error) {
	return json.Marshal(r)
}

// MakeRegisterRequest creates a signed peer.PeerRecord as a register request
// and marshals this into bytes
func MakeRegisterRequest(providerID peer.ID, privateKey crypto.PrivKey, addrs []string, account string) ([]byte, error) {
	if len(addrs) == 0 {
		return nil, errors.New("missing address")
	}

	rec := &RegisterRequest{}
	rec.PeerID = providerID
	rec.Addrs = addrs
	rec.Seq = peer.TimestampSeq()
	if account != "" {
		rec.MinerAccount = account
	}

	return makeRequestEnvelop(rec, privateKey)
}

// ReadRegisterRequest unmarshals a account.PeerRequest from bytes, verifies the
// signature, and returns a peer.PeerRecord
func ReadRegisterRequest(data []byte) (*RegisterRequest, error) {
	env, untypedRecord, err := record.ConsumeEnvelope(data, RegisterRequestEnvelopeDomain)
	if err != nil {
		return nil, fmt.Errorf("cannot consume register request envelope: %s", err)
	}
	rec, ok := untypedRecord.(*RegisterRequest)
	if !ok {
		return nil, fmt.Errorf("unmarshaled register request record is not a *PeerRecord")
	}
	isSendBySelf := rec.PeerID.MatchesPublicKey(env.PublicKey)
	if !isSendBySelf {
		return nil, fmt.Errorf("pubkey dismatch with peerid")
	}
	return rec, nil
}
