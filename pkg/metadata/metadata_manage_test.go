package metadata_test

import (
	"context"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/legs"
	. "github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/test/mock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestReceiveRecordAndOutUpdate(t *testing.T) {
	Convey("test metadata manager", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		err = logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		lys := legs.MkLinkSystem(pando.PS, nil, nil)

		Convey("give records when wait for maxInterval then update and backup", func() {
			//BackupMaxInterval = time.Second * 3
			pando.Opt.Backup.BackupGenInterval = (time.Second * 3).String()
			mm, err := New(context.Background(), pando.DS, &lys, pando.Registry, &pando.Opt.Backup)
			So(err, ShouldBeNil)
			provider, err := mock.NewMockProvider(pando)
			So(err, ShouldBeNil)
			err = pando.Registry.RegisterOrUpdate(context.Background(), provider.ID, cid.Undef, provider.ID, false)
			So(err, ShouldBeNil)
			cid1, err := provider.SendMeta(true)
			So(err, ShouldBeNil)
			cid2, err := provider.SendMeta(true)
			So(err, ShouldBeNil)
			cid3, err := provider.SendMeta(true)
			So(err, ShouldBeNil)
			mockRecord := []*MetaRecord{
				{cid1, provider.ID, uint64(time.Now().UnixNano())},
				{cid2, provider.ID, uint64(time.Now().UnixNano())},
				{cid3, provider.ID, uint64(time.Now().UnixNano())},
			}
			recvCh := mm.GetMetaInCh()
			for _, r := range mockRecord {
				recvCh <- r
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			t.Cleanup(func() {
				cancel()
			})

			time.Sleep(time.Second * 5)
			update, _, err := pando.PS.SnapShotStore.GetSnapShotByHeight(ctx, 0)
			So(err, ShouldBeNil)
			So(update.PrevSnapShot, ShouldEqual, "")
			So(len(update.Update[provider.ID.String()].MetaList), ShouldEqual, 3)
			So(update.Update[provider.ID.String()].MetaList, ShouldContain, cid1)
			So(update.Update[provider.ID.String()].MetaList, ShouldContain, cid2)
			So(update.Update[provider.ID.String()].MetaList, ShouldContain, cid3)
			t.Logf("%#v", update)

			mm.Close()
		})
	})
}
