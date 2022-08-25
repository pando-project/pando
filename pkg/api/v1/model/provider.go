package model

import (
	"encoding/json"
	"fmt"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/record"
)

func makeRequestEnvelop(rec record.Record, privateKey crypto.PrivKey) ([]byte, error) {
	envelope, err := record.Seal(rec, privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not sign request: %s", err)
	}

	data, err := envelope.Marshal()
	if err != nil {
		return nil, fmt.Errorf("could not marshal request register: %s", err)
	}

	return data, nil
}

type RegisterRequest struct {
	// PeerID is the ID of the account this record pertains to.
	PeerID peer.ID

	// Addrs contains the public addresses of the account this record pertains to.
	Addrs []string

	Name string

	// Seq is a monotonically-increasing sequence counter that's used to order
	// PeerRecords in time. The interval between Seq values is unspecified,
	// but newer PeerRecords MUST have a greater Seq value than older records
	// for the same account.
	Seq uint64

	// filecoin miner account
	MinerAccount string
}

const RequestEnvelopeDomain = "pando-register-request-record"

var RequestEnvelopePayloadType = []byte("pando-register-request")

//
func init() {
	record.RegisterType(&RegisterRequest{})
}

// Domain is used when signing and validating IngestRequest records contained in Envelopes
func (r *RegisterRequest) Domain() string {
	return RequestEnvelopeDomain
}

// Codec is a binary identifier for the IngestRequest types
func (r *RegisterRequest) Codec() []byte {
	return RequestEnvelopePayloadType
}

// UnmarshalRecord parses an IngestRequest from a byte slice.
func (r *RegisterRequest) UnmarshalRecord(data []byte) error {
	if r == nil {
		return fmt.Errorf("cannot unmarshal IngestRequest to nil receiver")
	}

	return json.Unmarshal(data, r)
}

// MarshalRecord serializes an IngestRequest to a byte slice.
func (r *RegisterRequest) MarshalRecord() ([]byte, error) {
	return json.Marshal(r)
}

// MakeRegisterRequest creates a signed peer.PeerRecord as a register request
// and marshals this into bytes
func MakeRegisterRequest(providerID peer.ID, privateKey crypto.PrivKey, addrs []string, account string, name string) ([]byte, error) {

	rec := &RegisterRequest{}
	rec.PeerID = providerID
	rec.Addrs = addrs
	rec.Seq = peer.TimestampSeq()
	rec.MinerAccount = account
	rec.Name = name

	return makeRequestEnvelop(rec, privateKey)
}

// ReadRegisterRequest unmarshal an account.PeerRequest from bytes, verifies the
// signature, and returns a peer.PeerRecord
func ReadRegisterRequest(data []byte) (*RegisterRequest, error) {
	env, untypedRecord, err := record.ConsumeEnvelope(data, RequestEnvelopeDomain)
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

type providerInfoRes map[string]struct {
	MultiAddr []string
	MinerAddr string
}

func GetProviderRes(info []*registry.ProviderInfo) ([]byte, error) {
	res := make(map[string]providerInfoRes)
	res["registeredProviders"] = make(providerInfoRes)
	provInfos := res["registeredProviders"]
	for _, provider := range info {
		peeridStr := provider.AddrInfo.ID.String()
		addrs := make([]string, 0)
		for _, addr := range provider.AddrInfo.Addrs {
			addrs = append(addrs, addr.String())
		}
		provInfos[peeridStr] = struct {
			MultiAddr []string
			MinerAddr string
		}{MultiAddr: addrs, MinerAddr: provider.DiscoveryAddr}
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return resBytes, nil
}
