package registry_test

import (
	"context"
	"fmt"
	dataStoreFactory "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"pando/pkg/lotus"
	"pando/pkg/option"
	"pando/pkg/registry"
	. "pando/pkg/registry"
	"pando/pkg/registry/internal/syserr"
	"pando/test/mock"
	"testing"
	"time"
)

const (
	exceptID   = "12D3KooWK7CTS7cyWi51PeNE3cTjS2F2kDCZaQVU4A5xBmb9J1do"
	trustedID  = "12D3KooWSG3JuvEjRkSxt93ADTjQxqe4ExbBwSkQ9Zyk1WfBaZJF"
	trustedID2 = "12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr"

	minerDiscoAddr = "stitest999999"
	minerAddr      = "/ip4/127.0.0.1/tcp/9999"
	minerAddr2     = "/ip4/127.0.0.2/tcp/9999"
)

func TestNewRegistryAndClose(t *testing.T) {
	Convey("TestNewRegistryAndClose", t, func() {
		Convey("return nil, err if config is nil when new registry", func() {
			opt := option.New(nil)
			reg, err := NewRegistry(context.Background(), nil,
				&opt.AccountLevel,
				&dataStoreFactory.Datastore{},
				&lotus.Discoverer{})
			So(err, ShouldResemble, fmt.Errorf("nil config"))
			So(reg, ShouldBeNil)
		})

		Convey("test create and close register with discovery", func() {
			pando, err := mock.NewPandoMock()
			So(err, ShouldBeNil)
			r := pando.Registry
			err = r.Close()
			So(err, ShouldBeNil)
			err = r.Close()
			So(err, ShouldBeNil)
		})
	})
}

func TestRegisterAndDiscovery(t *testing.T) {
	Convey("test register and discovery", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		r := pando.Registry

		peerID, err := peer.Decode(trustedID)
		So(err, ShouldBeNil)
		blackID, _ := peer.Decode("12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr")
		maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3002")
		So(err, ShouldBeNil)

		registerCases := []struct {
			registerInfo *ProviderInfo
			expected     error
		}{
			{
				registerInfo: &ProviderInfo{
					AddrInfo: peer.AddrInfo{
						ID:    peerID,
						Addrs: []multiaddr.Multiaddr{maddr},
					},
					DiscoveryAddr: minerDiscoAddr,
				},
				expected: nil,
			},
			{
				registerInfo: &ProviderInfo{
					AddrInfo: peer.AddrInfo{
						ID:    blackID,
						Addrs: []multiaddr.Multiaddr{maddr},
					},
				},
				expected: syserr.New(ErrNotAllowed, http.StatusForbidden),
			},
		}

		Convey("register, get provider info and reload", func() {
			ctx := context.Background()
			for _, tt := range registerCases {
				res := r.Register(ctx, tt.registerInfo)
				So(res, ShouldResemble, tt.expected)
			}
			infos := r.AllProviderInfo()
			So(len(infos), ShouldEqual, 1)
			l, err := r.ProviderAccountLevel(peerID)
			So(err, ShouldBeNil)
			So(l, ShouldEqual, 2)
			time.Sleep(time.Second * 2)
			isregister := r.IsRegistered(peerID)
			So(isregister, ShouldBeTrue)
			isregister = r.IsRegistered("?????")
			So(isregister, ShouldBeFalse)
			info := r.ProviderInfoByAddr(minerDiscoAddr)
			So(info, ShouldNotBeNil)
			err = r.Close()
			So(err, ShouldBeNil)

			diso, err := mock.NewMockDiscoverer(peerID.String())
			So(err, ShouldBeNil)
			// reload the persisted info
			r, err := registry.NewRegistry(ctx, &mock.MockDiscoveryCfg, &mock.MockAclCfg, pando.DS, diso)
			So(err, ShouldBeNil)
			info = r.ProviderInfo(peerID)
			So(info, ShouldNotBeNil)

		})

	})
}
