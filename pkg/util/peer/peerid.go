package peer

import (
	"encoding/base64"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func GetPeerIDFromPrivateKeyStr(privateKeyStr string) (peer.ID, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", err
	}

	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		return "", err
	}

	peerID, err := peer.IDFromPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	return peerID, nil
}
