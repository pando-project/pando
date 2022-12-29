package lotus

import (
	"context"
	"math/big"
	"testing"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pando-project/pando/pkg/registry"
	"github.com/pando-project/pando/pkg/registry/discovery"
	"github.com/stretchr/testify/assert"
)

const testMinerAddr = "t01000"

func TestDiscoverMock(t *testing.T) {
	t.Run("test discovery test", func(t *testing.T) {
		peerID, err := peer.Decode("12D3KooWRqmtFv7ccFfjR7RDcevoMEMXdCHNR8JNN8aNiH2dgk8Z")
		asserts := assert.New(t)
		asserts.Nil(err)
		diso, err := NewDiscoverer("???")
		asserts.Nil(err)
		data, err := diso.Discover(context.Background(), peerID, testMinerAddr)
		asserts.Nil(err)
		asserts.Equal(&discovery.Discovered{
			AddrInfo: peer.AddrInfo{ID: peerID},
			Type:     discovery.MinerType,
			Balance:  big.NewInt(0).Mul(registry.FIL, big.NewInt(5)),
		}, data)
	})
}

//func TestDiscover(t *testing.T) {
//	Convey("test get peer addrInfo from miner account", t, func() {
//		peerID, err := peer.Decode("12D3KooWGuQafP1HDkE2ixXZnX6q6LLygsUG1uoxaQEtfPAt5ygp")
//		asserts.Nil(err)
//		// Now use the default url, not the gateway input
//		diso, err := NewDiscoverer("???")
//		asserts.Nil(err)
//		data, err := diso._Discover(context.Background(), peerID, testMinerAddr)
//		asserts.Nil(err)
//		So(data.AddrInfo.ID, ShouldEqual, peerID)
//	})
//}
