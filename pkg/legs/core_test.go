package legs_test

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando-store/pkg/config"
	"github.com/kenlabs/pando-store/pkg/store"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/test/mock"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var _ = logging.SetLogLevel("*", "error")
var _ = logging.SetLogLevel("core", "debug")

func TestCreate(t *testing.T) {
	Convey("test create legs core", t, func() {
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		c := p.Core
		err = c.Close()
		So(err, ShouldBeNil)
		err = c.Close()
		So(err, ShouldBeNil)
	})

}

func TestGetMetaRecord(t *testing.T) {
	Convey("test get meta record", t, func() {
		_ = logging.SetLogLevel("*", "warn")
		ctx, cncl := context.WithTimeout(context.Background(), time.Second*10)
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		//core := p.Core
		outCh, err := p.GetMetaRecordCh()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(p)
		So(err, ShouldBeNil)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 5)
		cid, err := provider.SendMeta(true)
		So(err, ShouldBeNil)
		select {
		case <-ctx.Done():
			t.Error("timeout")
		case r := <-outCh:
			So(r.Cid, ShouldResemble, cid)
			So(r.ProviderID, ShouldResemble, provider.ID)
		}

		opt := option.New(nil)
		_, err = opt.Parse()
		if err != nil {
			t.Error(err)
		}
		c, err := legs.NewLegsCore(ctx, p.Host, p.DS, p.CS, p.PS, nil, time.Minute, nil, p.Registry, opt)
		So(err, ShouldBeNil)
		err = c.Close()
		So(err, ShouldBeNil)

		t.Cleanup(func() {
			cncl()
		})
	})
}

func TestRateLimiter(t *testing.T) {
	Convey("Test rate limiter", t, func() {
		patch := gomonkey.ApplyGlobalVar(&mock.BaseTokenRate, float64(1))
		defer patch.Reset()
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(p)
		So(err, ShouldBeNil)

		for i := 0; i < 1000; i++ {
			_, err = provider.SendDag()
			So(err, ShouldBeNil)
		}

		time.Sleep(time.Second * 10)

	})
}

func TestRecurseFetchMeta(t *testing.T) {
	Convey("Test fetch metadata recursely by golegs", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		ch, err := pando.GetMetaRecordCh()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(pando)
		So(err, ShouldBeNil)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(false)
			cids = append(cids, c)
			So(err, ShouldBeNil)
			t.Logf("send meta[cid:%s]", c.String())
		}
		time.Sleep(time.Second)
		ctx, cncl := context.WithTimeout(context.Background(), time.Minute)
		t.Cleanup(
			cncl,
		)

		c, err := provider.SendMeta(true)
		So(err, ShouldBeNil)
		cids = append(cids, c)

		time.Sleep(time.Second * 5)
		for i := 0; i < 6; i++ {
			_, err = pando.PS.Get(ctx, cids[i])
			So(err, ShouldBeNil)
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
	Convey("Test sync meta data from Pando by golegs", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(pando)
		So(err, ShouldBeNil)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(true)
			cids = append(cids, c)
			So(err, ShouldBeNil)
		}
		time.Sleep(time.Second)

		h, err := libp2p.New()
		So(err, ShouldBeNil)
		ds := datastore.NewMapDatastore()
		mds := sync.MutexWrap(ds)
		////bs := blockstore.NewBlockstore(mds)
		ps, err := store.NewStoreFromDatastore(context.Background(), mds, &config.StoreConfig{
			SnapShotInterval: "1s",
			CacheSize:        config.DefaultCacheSize,
		})
		So(err, ShouldBeNil)
		lsys := legs.MkLinkSystem(ps, nil, nil)
		consumer, err := golegs.NewSubscriber(h, mds, lsys, mock.GetTopic(), nil)
		So(err, ShouldBeNil)

		multiAddress := pando.Host.Addrs()[0].String() + "/ipfs/" + pando.Host.ID().String()
		peerInfo, err := peer.AddrInfoFromString(multiAddress)
		So(err, ShouldBeNil)
		err = h.Connect(context.Background(), *peerInfo)
		So(err, ShouldBeNil)

		c, err := consumer.Sync(context.Background(), pando.Host.ID(), cids[4], nil, nil)
		So(err, ShouldBeNil)
		So(c.Equals(cids[4]), ShouldBeTrue)

		for i := 0; i < 5; i++ {
			_, err := ps.Get(context.Background(), cids[i])
			So(err, ShouldBeNil)
		}

	})
}

func TestGetPayloadFromLink(t *testing.T) {
	Convey("Test fetch payload data from link", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		ch, err := pando.GetMetaRecordCh()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(pando)
		So(err, ShouldBeNil)
		var cids []cid.Cid
		var payloadCids []cid.Cid
		for i := 0; i < 5; i++ {
			c, pc, err := provider.SendMetaWithDataLink(true)
			cids = append(cids, c)
			payloadCids = append(payloadCids, pc)
			So(err, ShouldBeNil)
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
			So(err, ShouldBeNil)
			_, err = pando.PS.Get(ctx, payloadCids[i])
			So(err, ShouldBeNil)
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
