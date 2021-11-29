package config

import (
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestIdentity(t *testing.T) {
	Convey("test create and decode identity", t, func() {
		id, err := CreateIdentity(io.Discard)
		So(err, ShouldBeNil)
		pk, err := id.DecodePrivateKey("")
		So(err, ShouldBeNil)
		So(pk.Type(), ShouldEqual, crypto_pb.KeyType_Ed25519)
	})
}
