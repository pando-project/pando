package legs_test

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"pando/pkg/legs"
	"pando/test/mock"
	"testing"
	"time"
)

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
		core := p.Core
		outCh, err := p.GetMetaRecordCh()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(p)
		So(err, ShouldBeNil)
		//err = core.Subscribe(context.Background(), provider.ID)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 2)
		//cidlist, err := provider.SendDag()
		cid, err := provider.SendMeta()
		So(err, ShouldBeNil)
		select {
		case <-ctx.Done():
			t.Error("timeout")
		case r := <-outCh:
			So(r.Cid, ShouldResemble, cid)
			So(r.ProviderID, ShouldResemble, provider.ID)
		}

		core_, err := legs.NewLegsCore(ctx, p.Host, p.DS, p.BS, nil, nil, p.Registry)
		So(err, ShouldBeNil)

		t.Cleanup(func() {
			cncl()
			if err := provider.Close(); err != nil {
				t.Error(err)
			}
			if err := core.Close(); err != nil {
				t.Error(err)
			}
			if err := core_.Close(); err != nil {
				t.Error(err)
			}

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
