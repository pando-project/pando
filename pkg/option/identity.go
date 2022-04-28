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

	privKey, err := i.DecodePrivateKey()
	if err != nil {
		return "", nil, fmt.Errorf("could not decode private key: %w", err)
	}

	peerIDFromPrivKey, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return "", nil, fmt.Errorf("could not generate peer ID from private key: %w", err)
	}

	if i.PeerID != "" {
		peerID, err := peer.Decode(i.PeerID)
		if err != nil {
			return "", nil, fmt.Errorf("could not decode peer id: %w", err)
		}

		if peerID != "" && peerIDFromPrivKey != peerID {
			return "", nil, fmt.Errorf("provided peer ID must either match the peer ID generated from private key or be omitted: expected %s but got %s", peerIDFromPrivKey, peerID)
		}
	}

	return peerIDFromPrivKey, privKey, nil
}
