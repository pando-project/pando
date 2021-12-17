package registry_test

import (
	"Pando/internal/registry"
	. "Pando/internal/registry"
	"Pando/internal/syserr"
	"Pando/test/mock"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
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
	Convey("test create and close register with discovery", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		r := pando.Registry
		err = r.Close()
		So(err, ShouldBeNil)
		err = r.Close()
		So(err, ShouldBeNil)
	})
}

func TestRegisterAndDiscovery(t *testing.T) {
	Convey("test register and discovery", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		r := pando.Registry

		peerID, err := peer.Decode(trustedID)
		So(err, ShouldBeNil)
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
						ID:    "dasdsada",
						Addrs: []multiaddr.Multiaddr{maddr},
					},
				},
				expected: syserr.New(ErrNotAllowed, http.StatusForbidden),
			},
		}
		Convey("register, get provider info and reload", func() {
			for _, tt := range registerCases {
				res := r.Register(tt.registerInfo)
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
			r, err := registry.NewRegistry(&mock.MockDiscoveryCfg, &mock.MockAclCfg, pando.DS, diso, nil)
			So(err, ShouldBeNil)
			info = r.ProviderInfo(peerID)
			So(info, ShouldNotBeNil)

		})

	})
}

//func TestDiscoveryAllowed(t *testing.T) {
//	mockDisco := newMockDiscoverer(t, exceptID)
//
//	r, err := NewRegistry(&discoveryCfg, &aclCfg, nil, mockDisco)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer r.Close()
//	t.Log("created new registry")
//
//	peerID, err := peer.Decode(exceptID)
//	if err != nil {
//		t.Fatal("bad provider ID:", err)
//	}
//
//	err = r.Discover(peerID, minerDiscoAddr, true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Log("discovered mock miner", minerDiscoAddr)
//
//	info := r.ProviderInfoByAddr(minerDiscoAddr)
//	if info == nil {
//		t.Fatal("did not get provider info for miner")
//	}
//	t.Log("got provider info for miner")
//
//	assert.Equal(t, info.AddrInfo.ID, peerID, "did not get correct porvider id")
//
//	peerID, err = peer.Decode(trustedID)
//	if err != nil {
//		t.Fatal("bad provider ID:", err)
//	}
//	maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3002")
//	if err != nil {
//		t.Fatalf("Cannot create multiaddr: %s", err)
//	}
//	info = &ProviderInfo{
//		AddrInfo: peer.AddrInfo{
//			ID:    peerID,
//			Addrs: []multiaddr.Multiaddr{maddr},
//		},
//	}
//	err = r.Register(info)
//	if err != nil {
//		t.Error("failed to register directly:", err)
//	}
//
//	infos := r.AllProviderInfo()
//	if len(infos) != 2 {
//		t.Fatal("expected 2 provider infos")
//	}
//
//}
//
//func TestDiscoveryBlocked(t *testing.T) {
//	mockDisco := newMockDiscoverer(t, exceptID)
//
//	peerID, err := peer.Decode(exceptID)
//	if err != nil {
//		t.Fatal("bad provider ID:", err)
//	}
//
//	discoveryCfg.Policy.Allow = true
//	defer func() {
//		discoveryCfg.Policy.Allow = false
//	}()
//
//	r, err := NewRegistry(&discoveryCfg, &aclCfg, nil, mockDisco)
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer r.Close()
//	t.Log("created new registry")
//
//	err = r.Discover(peerID, minerDiscoAddr, true)
//	if !errors.Is(err, ErrNotAllowed) {
//		t.Fatal("expected error:", ErrNotAllowed, "got:", err)
//	}
//
//	into := r.ProviderInfoByAddr(minerDiscoAddr)
//	if into != nil {
//		t.Error("should not have found provider info for miner")
//	}
//}
//
//func TestDatastore(t *testing.T) {
//	dataStorePath := t.TempDir()
//	mockDisco := newMockDiscoverer(t, exceptID)
//
//	peerID, err := peer.Decode(trustedID)
//	if err != nil {
//		t.Fatal("bad provider ID:", err)
//	}
//	maddr, err := multiaddr.NewMultiaddr(minerAddr)
//	if err != nil {
//		t.Fatal("bad miner address:", err)
//	}
//	info1 := &ProviderInfo{
//		AddrInfo: peer.AddrInfo{
//			ID:    peerID,
//			Addrs: []multiaddr.Multiaddr{maddr},
//		},
//	}
//	peerID, err = peer.Decode(trustedID2)
//	if err != nil {
//		t.Fatal("bad provider ID:", err)
//	}
//	maddr, err = multiaddr.NewMultiaddr(minerAddr2)
//	if err != nil {
//		t.Fatal("bad miner address:", err)
//	}
//	info2 := &ProviderInfo{
//		AddrInfo: peer.AddrInfo{
//			ID:    peerID,
//			Addrs: []multiaddr.Multiaddr{maddr},
//		},
//	}
//
//	// Create datastore
//	dstore, err := leveldb.NewDatastore(dataStorePath, nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	r, err := NewRegistry(&discoveryCfg, &aclCfg, dstore, mockDisco)
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Log("created new registry with datastore")
//
//	err = r.Register(info1)
//	if err != nil {
//		t.Fatal("failed to register directly:", err)
//	}
//	err = r.Register(info2)
//	if err != nil {
//		t.Fatal("failed to register directly:", err)
//	}
//
//	pinfo := r.ProviderInfo(peerID)
//	if pinfo == nil {
//		t.Fatal("did not find registered provider")
//	}
//
//	err = r.Close()
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Create datastore
//	dstore, err = leveldb.NewDatastore(dataStorePath, nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//	r, err = NewRegistry(&discoveryCfg, &aclCfg, dstore, mockDisco)
//	if err != nil {
//		t.Fatal(err)
//	}
//	t.Log("re-created new registry with datastore")
//
//	infos := r.AllProviderInfo()
//	if len(infos) != 2 {
//		t.Fatal("expected 2 provider infos")
//	}
//
//	for i := range infos {
//		pid := infos[i].AddrInfo.ID
//		if pid != info1.AddrInfo.ID && pid != info2.AddrInfo.ID {
//			t.Fatalf("loaded invalid provider ID: %s", pid)
//		}
//	}
//
//	err = r.Close()
//	if err != nil {
//		t.Fatal(err)
//	}
//}
