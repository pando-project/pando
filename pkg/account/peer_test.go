package account

import (
	"github.com/agiledragon/gomonkey/v2"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestFetchPeerType(t *testing.T) {
	Convey("TestFetchPeerType", t, func() {
		r := &registry.Registry{}

		Convey("When peer is registered", func() {
			patch := gomonkey.ApplyMethodFunc(
				reflect.TypeOf(r),
				"ProviderAccountLevel",
				func(_ peer.ID) (int, error) {
					return 1, nil
				},
			)

			Convey("Given an untrusted account level equals to 1, should returns registeredPeer type", func() {
				patch = patch.ApplyMethodFunc(
					reflect.TypeOf(r),
					"IsTrusted",
					func(_ peer.ID) bool {
						return false
					},
				)
				defer patch.Reset()

				providerInfo := FetchPeerType("", r)
				So(providerInfo.PeerType, ShouldEqual, RegisteredPeer)
				So(providerInfo.AccountLevel, ShouldEqual, 1)
			})

			Convey("Given a trusted account, should returns whitelistPeer type", func() {
				patch = patch.ApplyMethodFunc(
					reflect.TypeOf(r),
					"IsTrusted",
					func(_ peer.ID) bool {
						return true
					},
				)
				defer patch.Reset()

				providerInfo := FetchPeerType("", r)
				So(providerInfo.PeerType, ShouldEqual, WhiteListPeer)
				So(providerInfo.AccountLevel, ShouldEqual, 1)
			})
		})

		Convey("when peer is not registered", func() {
			patch := gomonkey.ApplyMethodFunc(
				reflect.TypeOf(r),
				"ProviderAccountLevel",
				func(_ peer.ID) (int, error) {
					return -1, nil
				},
			)
			patch = patch.ApplyMethodFunc(
				reflect.TypeOf(r),
				"IsTrusted",
				func(_ peer.ID) bool {
					return false
				},
			)
			defer patch.Reset()

			providerInfo := FetchPeerType("", r)
			So(providerInfo.PeerType, ShouldEqual, UnregisteredPeer)
			So(providerInfo.AccountLevel, ShouldEqual, -1)
		})

	})
}
