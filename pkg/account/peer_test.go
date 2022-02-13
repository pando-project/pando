package account

import (
	"github.com/agiledragon/gomonkey/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/smartystreets/goconvey/convey"
	"pando/pkg/registry"
	"reflect"
	"testing"
)

func TestFetchPeerType(t *testing.T) {
	convey.Convey("when give peer id then get peer type", t, func() {
		r := &registry.Registry{}
		patch1 := gomonkey.ApplyMethod(reflect.TypeOf(r), "ProviderAccountLevel", func(_ *registry.Registry, _ peer.ID) (int, error) {
			return 1, nil
		})
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(r), "IsTrusted", func(_ *registry.Registry, _ peer.ID) bool {
			return false
		})
		defer patch1.Reset()
		defer patch2.Reset()

		info := FetchPeerType(peer.ID(""), r)
		convey.So(info.PeerType, convey.ShouldEqual, RegisteredPeer)
		convey.So(info.AccountLevel, convey.ShouldEqual, 1)

		patch2 = gomonkey.ApplyMethod(reflect.TypeOf(r), "IsTrusted", func(_ *registry.Registry, _ peer.ID) bool {
			return true
		})

		info = FetchPeerType(peer.ID(""), r)
		convey.So(info.PeerType, convey.ShouldEqual, WhiteListPeer)
	})
}
