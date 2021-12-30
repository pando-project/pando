package option

import (
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/libp2p/go-libp2p-core/crypto"
)

type Identity struct {
	PeerID     string `yaml:"PeerID"`
	PrivateKey string `yaml:"PrivateKey"`
}

func (i Identity) DecodePrivateKey() (crypto.PrivKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(i.PrivateKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(privateKeyBytes)
}

func (i Identity) Decode() (peer.ID, crypto.PrivKey, error) {
	peerID, err := peer.Decode(i.PeerID)
	if err != nil {
		return "", nil, fmt.Errorf("could not decode account id: %s", err)
	}

	privateKey, err := i.DecodePrivateKey()
	if err != nil {
		return "", nil, fmt.Errorf("could not decode private key: %s", err)
	}

	return peerID, privateKey, nil
}
