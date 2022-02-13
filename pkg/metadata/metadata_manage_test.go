package metadata_test

import (
	"context"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"pando/pkg/legs"
	. "pando/pkg/metadata"
	"pando/test/mock"
	"testing"
	"time"
)

func TestReceiveRecordAndOutUpdate(t *testing.T) {
	Convey("test metadata manager", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		err = logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		lys := legs.MkLinkSystem(pando.BS, nil)

		Convey("give records when wait for maxInterval then update and backup", func() {
			//BackupMaxInterval = time.Second * 3
			pando.Opt.Backup.BackupGenInterval = (time.Second * 3).String()
			mm, err := New(context.Background(), pando.DS, pando.BS, &lys, pando.Registry, &pando.Opt.Backup)
			So(err, ShouldBeNil)
			provider, err := mock.NewMockProvider(pando)
			So(err, ShouldBeNil)
			err = pando.Registry.RegisterOrUpdate(context.Background(), provider.ID, cid.Undef)
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
			outCh := mm.GetUpdateOut()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			t.Cleanup(func() {
				cancel()
			})

			time.Sleep(time.Second * 5)
			select {
			case <-ctx.Done():
				t.Error("timeout!not get update rightly")
			case update := <-outCh:
				So(len(update), ShouldEqual, 1)
				So(update, ShouldContainKey, mockRecord[0].ProviderID)
				So(update[mockRecord[0].ProviderID].Cidlist, ShouldResemble, []cid.Cid{mockRecord[0].Cid, mockRecord[1].Cid, mockRecord[2].Cid})
			}
			mm.Close()
		})
	})
}
