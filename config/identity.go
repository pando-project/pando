package config

import (
	"encoding/base64"
	ic "github.com/libp2p/go-libp2p-core/crypto"
)

// Identity tracks the configuration of the local node's identity.
type Identity struct {
	PeerID  string
	PrivKey string `json:",omitempty"`
}

// DecodePrivateKey is a helper to decode the user's PrivateKey.
func (i Identity) DecodePrivateKey(passphrase string) (ic.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}
	return ic.UnmarshalPrivateKey(pkb)
}
