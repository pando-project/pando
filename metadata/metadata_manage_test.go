package metadata_test

import (
	. "Pando/metadata"
	"Pando/test/mock"
	"context"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
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

		Convey("test receive record, out update and backup because of maxInterval", func() {
			BackupMaxInterval = time.Second * time.Second * 3
			mm, err := New(context.Background(), pando.DS, pando.BS)
			So(err, ShouldBeNil)
			provider, err := mock.NewMockProvider(pando)
			So(err, ShouldBeNil)
			err = pando.Core.Subscribe(context.Background(), provider.ID)
			So(err, ShouldBeNil)
			cidlist, err := provider.SendDag()
			So(err, ShouldBeNil)
			cidlist2, err := provider.SendDag()
			So(err, ShouldBeNil)
			cidlist3, err := provider.SendDag()
			So(err, ShouldBeNil)
			mockRecord := []*MetaRecord{
				{cidlist[0], provider.ID, uint64(time.Now().UnixNano())},
				{cidlist2[0], provider.ID, uint64(time.Now().UnixNano())},
				{cidlist3[0], provider.ID, uint64(time.Now().UnixNano())},
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
		Convey("test receive record, out update and backup because of maxDagNum", func() {
			BackupMaxInterval = time.Second * 60
			BackupCheckNumInterval = time.Second
			BackupMaxDagNums = 1
			mm, err := New(context.Background(), pando.DS, pando.BS)
			So(err, ShouldBeNil)
			provider, err := mock.NewMockProvider(pando)
			So(err, ShouldBeNil)
			err = pando.Core.Subscribe(context.Background(), provider.ID)
			So(err, ShouldBeNil)
			cidlist, err := provider.SendDag()
			So(err, ShouldBeNil)
			cidlist2, err := provider.SendDag()
			So(err, ShouldBeNil)
			cidlist3, err := provider.SendDag()
			So(err, ShouldBeNil)
			mockRecord := []*MetaRecord{
				{cidlist[0], provider.ID, uint64(time.Now().UnixNano())},
				{cidlist2[0], provider.ID, uint64(time.Now().UnixNano())},
				{cidlist3[0], provider.ID, uint64(time.Now().UnixNano())},
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

			select {
			case <-ctx.Done():
				t.Error("timeout!not get update rightly")
			case update := <-outCh:
				So(len(update), ShouldEqual, 1)
				So(update, ShouldContainKey, mockRecord[0].ProviderID)
				So(update[mockRecord[0].ProviderID].Cidlist, ShouldResemble, []cid.Cid{mockRecord[0].Cid, mockRecord[1].Cid, mockRecord[2].Cid})

			}
		})

	})
}
