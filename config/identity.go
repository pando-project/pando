package config

import (
	"encoding/base64"
	"fmt"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
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

func (i Identity) Decode() (peer.ID, ic.PrivKey, error) {
	peerID, err := peer.Decode(i.PeerID)
	if err != nil {
		return "", nil, fmt.Errorf("could not decode account id: %s", err)
	}

	privKey, err := i.DecodePrivateKey("")
	if err != nil {
		return "", nil, fmt.Errorf("could not decode private key: %s", err)
	}

	return peerID, privKey, nil
}
