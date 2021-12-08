package legs_test

import (
	"Pando/test/mock"
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

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
		cidlist, err := provider.SendDag()
		So(err, ShouldBeNil)
		select {
		case _ = <-ctx.Done():
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
