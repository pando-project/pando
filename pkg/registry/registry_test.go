package registry_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pando-project/pando/pkg/lotus"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/registry"
	. "github.com/pando-project/pando/pkg/registry"
	"github.com/pando-project/pando/pkg/registry/internal/syserr"
	"github.com/pando-project/pando/test/mock"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dataStoreFactory "github.com/ipfs/go-ds-leveldb"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
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
	t.Run("TestNewRegistryAndClose", func(t *testing.T) {
		asserts := assert.New(t)
		t.Run("return nil, err if config is nil when new registry", func(t *testing.T) {
			opt := option.New(nil)
			reg, err := NewRegistry(context.Background(), nil,
				&opt.AccountLevel,
				&dataStoreFactory.Datastore{},
				&lotus.Discoverer{})
			asserts.Equal(fmt.Errorf("nil config"), err)
			asserts.Nil(reg)
		})

		t.Run("test create and close register with discovery", func(t *testing.T) {
			pando, err := mock.NewPandoMock()
			asserts.Nil(err)
			r := pando.Registry
			err = r.Close()
			asserts.Nil(err)
			err = r.Close()
			asserts.Nil(err)
		})
	})
}

func TestRegisterAndDiscovery(t *testing.T) {
	asserts := assert.New(t)
	t.Run("test register and discovery", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts.Nil(err)
		r := pando.Registry

		peerID, err := peer.Decode(trustedID)
		asserts.Nil(err)
		blackID, _ := peer.Decode("12D3KooWKSNuuq77xqnpPLnU3fq1bTQW2TwSZL2Z4QTHEYpUVzfr")
		maddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/3002")
		asserts.Nil(err)

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

		t.Run("register, get provider info and reload", func(t *testing.T) {
			ctx := context.Background()
			for _, tt := range registerCases {
				res := r.Register(ctx, tt.registerInfo)
				asserts.Equal(tt.expected, res)
			}
			infos := r.AllProviderInfo()
			asserts.Equal(1, len(infos))
			l, err := r.ProviderAccountLevel(peerID)
			asserts.Nil(err)
			asserts.Equal(2, l)
			time.Sleep(time.Second * 2)
			isregister := r.IsRegistered(peerID)
			asserts.True(isregister)
			isregister = r.IsRegistered("?????")
			asserts.False(isregister)
			info := r.ProviderInfoByAddr(minerDiscoAddr)
			asserts.NotNil(info)
			err = r.Close()
			asserts.Nil(err)

			diso, err := mock.NewMockDiscoverer(peerID.String())
			asserts.Nil(err)
			// reload the persisted info
			r, err := registry.NewRegistry(ctx, &mock.MockDiscoveryCfg, &mock.MockAclCfg, pando.DS, diso)
			asserts.Nil(err)
			info = r.ProviderInfo(peerID)[0]
			asserts.NotNil(info)
		})

	})
}

func TestAutoRegister(t *testing.T) {
	asserts := assert.New(t)
	t.Run("test register auto", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts.Nil(err)
		//outCh, err := pando.GetMetaRecordCh()
		provider, err := mock.NewMockProvider(pando)
		asserts.Nil(err)
		time.Sleep(time.Millisecond * 500)
		_, err = provider.SendDag()
		asserts.Nil(err)
		time.Sleep(time.Millisecond * 500)
		res := pando.Registry.ProviderInfo(provider.ID)
		asserts.Nil(res)
		c, err := provider.SendMeta(true)
		time.Sleep(time.Millisecond * 500)

		res = pando.Registry.ProviderInfo(provider.ID)
		asserts.NotNil(res)
		asserts.Equal(provider.ID.String(), res[0].AddrInfo.ID.String())
		lastSync, err := pando.Core.DS.Get(context.Background(), datastore.NewKey("/sync/"+res[0].AddrInfo.ID.String()))
		asserts.Nil(err)
		_, lastSyncCid, err := cid.CidFromBytes(lastSync)
		asserts.Nil(err)
		asserts.True(lastSyncCid.Equals(c))
		lastSyncCidLs := pando.Core.LS.GetLatestSync(provider.ID).(cidlink.Link).Cid
		asserts.True(lastSyncCidLs.Equals(c))
	})

}
