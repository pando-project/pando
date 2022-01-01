package system

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

func CreateIdentity() (PeerID string, PrivateKey string, err error) {
	fmt.Println("generating ED25519 keypair...")
	privateKey, publicKey, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return
	}

	privateKeyBytes, err := crypto.MarshalPrivateKey(privateKey)
	if err != nil {
		return
	}
	PrivateKey = base64.StdEncoding.EncodeToString(privateKeyBytes)

	peerID, err := peer.IDFromPublicKey(publicKey)
	if err != nil {
		return
	}
	PeerID = peerID.Pretty()

	fmt.Printf("keypair generated, peer id: %s\n", PeerID)
	return
}
