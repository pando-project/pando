package account

import (
	"testing"
)

func TestFetchPeerType(t *testing.T) {
	t.Run("TestFetchPeerType", func(t *testing.T) {
		//r := &registry.Registry{}

		t.Run("When peer is registered", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(r),
			//	"ProviderAccountLevel",
			//	func(_ peer.ID) (int, error) {
			//		return 1, nil
			//	},
			//)

			t.Run("Given an untrusted account level equals to 1, should returns registeredPeer type", func(t *testing.T) {
				//patch = patch.ApplyMethodFunc(
				//	reflect.TypeOf(r),
				//	"IsTrusted",
				//	func(_ peer.ID) bool {
				//		return false
				//	},
				//)
				//defer patch.Reset()
				//
				//providerInfo := FetchPeerType("", r)
				//So(providerInfo.PeerType, ShouldEqual, RegisteredPeer)
				//So(providerInfo.AccountLevel, ShouldEqual, 1)
			})

			t.Run("Given a trusted account, should returns whitelistPeer type", func(t *testing.T) {
				//patch = patch.ApplyMethodFunc(
				//	reflect.TypeOf(r),
				//	"IsTrusted",
				//	func(_ peer.ID) bool {
				//		return true
				//	},
				//)
				//defer patch.Reset()
				//
				//providerInfo := FetchPeerType("", r)
				//So(providerInfo.PeerType, ShouldEqual, WhiteListPeer)
				//So(providerInfo.AccountLevel, ShouldEqual, 1)
			})
		})

		t.Run("when peer is not registered", func(t *testing.T) {
			//patch := gomonkey.ApplyMethodFunc(
			//	reflect.TypeOf(r),
			//	"ProviderAccountLevel",
			//	func(_ peer.ID) (int, error) {
			//		return -1, nil
			//	},
			//)
			//patch = patch.ApplyMethodFunc(
			//	reflect.TypeOf(r),
			//	"IsTrusted",
			//	func(_ peer.ID) bool {
			//		return false
			//	},
			//)
			//defer patch.Reset()
			//
			//providerInfo := FetchPeerType("", r)
			//So(providerInfo.PeerType, ShouldEqual, UnregisteredPeer)
			//So(providerInfo.AccountLevel, ShouldEqual, -1)
		})

	})
}
