package config

import (
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestIdentity(t *testing.T) {
	Convey("test create and decode identity", t, func() {
		identity, err := CreateIdentity(io.Discard)
		So(err, ShouldBeNil)
		pk, err := identity.DecodePrivateKey("")
		So(err, ShouldBeNil)
		So(pk.Type(), ShouldEqual, crypto_pb.KeyType_Ed25519)
		id, pk2, err := identity.Decode()
		So(err, ShouldBeNil)
		So(id.String(), ShouldEqual, identity.PeerID)
		So(pk2, ShouldResemble, pk)
	})
}
