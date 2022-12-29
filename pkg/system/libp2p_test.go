package system

import (
	"testing"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

func TestCreateIdentity(t *testing.T) {
	t.Run("TestCreateIdentity", func(t *testing.T) {
		t.Run("return a peerID and a privateKey which should be matched with each other", func(t *testing.T) {
			peerIDStr, privateKeyStr, err := CreateIdentity()

			peerID, err := peer.Decode(peerIDStr)
			if err != nil {
				t.Error(err)
			}

			privateKeyBytes, err := crypto.ConfigDecodeKey(privateKeyStr)
			if err != nil {
				t.Error(err)
			}
			privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
			if err != nil {
				t.Error(err)
			}
			assert.True(t, peerID.MatchesPrivateKey(privateKey))
		})
	})
}
