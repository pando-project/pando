package types

import "github.com/ipfs/go-cid"

//go:generate cbor-gen-for ProviderState

type ProviderState struct {
	Cidlist []cid.Cid
}
