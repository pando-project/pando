package system

import (
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateIdentity(t *testing.T) {
	Convey("TestCreateIdentity", t, func() {
		Convey("return a peerID and a privateKey which should be matched with each other", func() {
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

			So(peerID.MatchesPrivateKey(privateKey), ShouldBeTrue)
		})
	})
}
