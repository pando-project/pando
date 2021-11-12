package config

import (
	crypto_pb "github.com/libp2p/go-libp2p-core/crypto/pb"
	"io"
	"testing"
)

func TestIdentity(t *testing.T) {
	id, err := CreateIdentity(io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	pk, err := id.DecodePrivateKey("")
	if err != nil {
		t.Fatal(err)
	}
	if pk.Type() != crypto_pb.KeyType_Ed25519 {
		t.Fatal("unexpected type:", pk.Type())
	}
}
