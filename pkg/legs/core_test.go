package legs_test

import (
	"context"
	"testing"
	"time"

	"github.com/pando-project/pando/pkg/legs"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/test/mock"

	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando-store/pkg/config"
	"github.com/kenlabs/pando-store/pkg/store"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
)

var _ = logging.SetLogLevel("*", "error")
var _ = logging.SetLogLevel("core", "debug")

func TestCreate(t *testing.T) {
	t.Run("test create legs core", func(t *testing.T) {
		p, err := mock.NewPandoMock()
		asserts := assert.New(t)

		asserts.Nil(err)
		c := p.Core
		err = c.Close()
		asserts.Nil(err)
		err = c.Close()
		asserts.Nil(err)
	})

}

func TestGetMetaRecord(t *testing.T) {
	t.Run("test get meta record", func(t *testing.T) {
		_ = logging.SetLogLevel("*", "warn")
		ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
		p, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)

		outCh, err := p.GetMetaRecordCh()
		asserts.Nil(err)
		provider, err := mock.NewMockProvider(p)
		asserts.Nil(err)

		time.Sleep(time.Second * 5)
		cidd, err := provider.SendMeta(true)
		asserts.Nil(err)
		select {
		case <-ctx.Done():
			t.Error("timeout")
		case r := <-outCh:
			asserts.Equal(cidd, r.Cid)
			asserts.Equal(provider.ID, r.ProviderID)
		}

		opt := option.New(nil)
		_, err = opt.Parse()
		if err != nil {
			t.Error(err)
		}
		c, err := legs.NewLegsCore(ctx, p.Host, p.DS, p.CS, p.PS, nil, time.Minute, nil, p.Registry, opt)
		asserts.Nil(err)
		err = c.Close()
		asserts.Nil(err)

		t.Cleanup(func() {
			cncl()
		})
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("Test rate limiter", func(t *testing.T) {
		//ToDo: remove rate limiter and related tests
		//patch := gomonkey.ApplyGlobalVar(&mock.BaseTokenRate, float64(1))
		//defer patch.Reset()
		//p, err := mock.NewPandoMock()
		//asserts.Nil(err)
		//provider, err := mock.NewMockProvider(p)
		//asserts.Nil(err)
		//
		//for i := 0; i < 1000; i++ {
		//	_, err = provider.SendDag()
		//	asserts.Nil(err)
		//}
		//
		//time.Sleep(time.Second * 10)

	})
}

func TestRecurseFetchMeta(t *testing.T) {
	t.Run("Test fetch metadata recursely by golegs", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)
		ch, err := pando.GetMetaRecordCh()
		asserts.Nil(err)
		provider, err := mock.NewMockProvider(pando)
		asserts.Nil(err)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(false)
			cids = append(cids, c)
			asserts.Nil(err)
			t.Logf("send meta[cid:%s]", c.String())
		}
		time.Sleep(time.Second)
		ctx, cncl := context.WithTimeout(context.Background(), time.Minute)
		t.Cleanup(
			cncl,
		)

		c, err := provider.SendMeta(true)
		asserts.Nil(err)
		cids = append(cids, c)

		time.Sleep(time.Second * 5)
		for i := 0; i < 6; i++ {
			_, err = pando.PS.Get(ctx, cids[i])
			asserts.Nil(err)
			select {
			case rec, ok := <-ch:
				if !ok {
					t.Error("error closed")
				}
				t.Log(rec)
			}
		}
	})

}

func TestSyncDataFromPando(t *testing.T) {
	t.Run("Test sync meta data from Pando by golegs", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)
		provider, err := mock.NewMockProvider(pando)
		asserts.Nil(err)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(true)
			cids = append(cids, c)
			asserts.Nil(err)
		}
		time.Sleep(time.Second)

		h, err := libp2p.New()
		asserts.Nil(err)
		ds := datastore.NewMapDatastore()
		mds := sync.MutexWrap(ds)
		////bs := blockstore.NewBlockstore(mds)
		ps, err := store.NewStoreFromDatastore(context.Background(), mds, &config.StoreConfig{
			SnapShotInterval: "1s",
			CacheSize:        config.DefaultCacheSize,
		})
		asserts.Nil(err)
		lsys := legs.MkLinkSystem(ps, nil, nil)
		consumer, err := golegs.NewSubscriber(h, mds, lsys, mock.GetTopic(), nil)
		asserts.Nil(err)

		multiAddress := pando.Host.Addrs()[0].String() + "/ipfs/" + pando.Host.ID().String()
		peerInfo, err := peer.AddrInfoFromString(multiAddress)
		asserts.Nil(err)
		err = h.Connect(context.Background(), *peerInfo)
		asserts.Nil(err)

		c, err := consumer.Sync(context.Background(), pando.Host.ID(), cids[4], nil, nil)
		asserts.Nil(err)
		asserts.True(c.Equals(cids[4]))

		for i := 0; i < 5; i++ {
			_, err := ps.Get(context.Background(), cids[i])
			asserts.Nil(err)
		}

	})
}

func TestGetPayloadFromLink(t *testing.T) {
	t.Run("Test fetch payload data from link", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)
		ch, err := pando.GetMetaRecordCh()
		asserts.Nil(err)
		provider, err := mock.NewMockProvider(pando)
		asserts.Nil(err)
		var cids []cid.Cid
		var payloadCids []cid.Cid
		for i := 0; i < 5; i++ {
			c, pc, err := provider.SendMetaWithDataLink(true)
			cids = append(cids, c)
			payloadCids = append(payloadCids, pc)
			asserts.Nil(err)
			t.Logf("send meta[cid:%s] with payload link: %s", c.String(), pc.String())
		}
		time.Sleep(time.Second)
		ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
		t.Cleanup(
			cncl,
		)

		time.Sleep(time.Second * 1)
		for i := 0; i < 5; i++ {
			_, err = pando.PS.Get(ctx, cids[i])
			asserts.Nil(err)
			_, err = pando.PS.Get(ctx, payloadCids[i])
			asserts.Nil(err)
			select {
			case rec, ok := <-ch:
				if !ok {
					t.Error("error closed")
				}
				t.Log(rec.Cid)
			}
		}
	})
}
