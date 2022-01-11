package legs_test

import (
	"context"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	. "pando/pkg/legs"
	"pando/test/mock"
	"reflect"
	"testing"
	"time"
)

var _ = logging.SetLogLevel("core", "debug")

func TestCreate(t *testing.T) {
	Convey("test create legs core", t, func() {
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		c := p.Core
		err = c.Close(context.Background())
		So(err, ShouldBeNil)
		err = c.Close(context.Background())
		So(err, ShouldBeNil)
	})

}

func TestGetMetaRecord(t *testing.T) {
	Convey("test get meta record", t, func() {
		ctx, cncl := context.WithTimeout(context.Background(), time.Minute*5)
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := p.Core
		outCh, err := p.GetMetaRecordCh()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(p)
		So(err, ShouldBeNil)
		err = core.Subscribe(context.Background(), provider.ID)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 2)
		cidlist, err := provider.SendDag()
		So(err, ShouldBeNil)
		select {
		case <-ctx.Done():
			t.Error("timeout")
		case r := <-outCh:
			So(r.Cid, ShouldResemble, cidlist[0])
			So(r.ProviderID, ShouldResemble, provider.ID)
		}

		t.Cleanup(func() {
			cncl()
			if err := provider.Close(); err != nil {
				t.Error(err)
			}
			if err := core.Close(context.Background()); err != nil {
				t.Error(err)
			}
		})
	})
}

func TestRepeatSubAndUnsub(t *testing.T) {
	Convey("when repeat subscribing and unsubscribing provider then get error", t, func() {
		ctx, cncl := context.WithTimeout(context.Background(), time.Minute*5)
		defer cncl()
		testPeer, _ := peer.Decode("12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4")
		testPeer2, _ := peer.Decode("12D3KooWNtUworDmrdBTBrLqeD5s26MLnpRX1QJGQ46HXaJVBXV4")
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := p.Core
		err = core.Subscribe(ctx, testPeer)
		So(err, ShouldBeNil)
		err = core.Subscribe(ctx, testPeer)
		So(err, ShouldBeNil)
		patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(core), "getLatestSync", func(_ *Core, peerID peer.ID) (cid.Cid, error) {
			return cid.Undef, fmt.Errorf("unknown error")
		})
		defer patch.Reset()
		err = core.Unsubscribe(ctx, testPeer)
		So(err, ShouldBeNil)

		err = core.Subscribe(ctx, testPeer)
		So(err, ShouldResemble, fmt.Errorf("unknown error"))

		err = core.Unsubscribe(ctx, testPeer2)
		So(err, ShouldBeNil)
	})

}

func TestRateLimiter(t *testing.T) {
	Convey("Test rate limiter", t, func() {
		patch := gomonkey.ApplyGlobalVar(&mock.BaseTokenRate, float64(1))
		defer patch.Reset()
		ctx, cncl := context.WithTimeout(context.Background(), time.Minute*5)
		defer cncl()
		p, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := p.Core
		provider, err := mock.NewMockProvider(p)
		So(err, ShouldBeNil)
		err = core.Subscribe(ctx, provider.ID)
		So(err, ShouldBeNil)

		for i := 0; i < 1000; i++ {
			_, err = provider.SendDag()
			So(err, ShouldBeNil)
		}

		time.Sleep(time.Second * 10)

	})
}
