package lotus

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"math/big"
	"pando/pkg/registry"
	"pando/pkg/registry/discovery"
	"testing"
)

const testMinerAddr = "t01000"

func TestDiscoverMock(t *testing.T) {
	Convey("test discovery mock", t, func() {
		peerID, err := peer.Decode("12D3KooWRqmtFv7ccFfjR7RDcevoMEMXdCHNR8JNN8aNiH2dgk8Z")
		So(err, ShouldBeNil)
		diso, err := NewDiscoverer("???")
		So(err, ShouldBeNil)
		data, err := diso.Discover(context.Background(), peerID, testMinerAddr)
		So(err, ShouldBeNil)
		So(data, ShouldResemble, &discovery.Discovered{
			AddrInfo: peer.AddrInfo{ID: peerID},
			Type:     discovery.MinerType,
			Balance:  big.NewInt(0).Mul(registry.FIL, big.NewInt(5)),
		})
	})

}

func TestDiscover(t *testing.T) {
	Convey("test get peer addrInfo from miner account", t, func() {
		peerID, err := peer.Decode("12D3KooWGuQafP1HDkE2ixXZnX6q6LLygsUG1uoxaQEtfPAt5ygp")
		So(err, ShouldBeNil)
		// Now use the default url, not the gateway input
		diso, err := NewDiscoverer("???")
		So(err, ShouldBeNil)
		data, err := diso._Discover(context.Background(), peerID, testMinerAddr)
		So(err, ShouldBeNil)
		So(data.AddrInfo.ID, ShouldEqual, peerID)
	})
}
