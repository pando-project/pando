package metadata_test

import (
	. "Pando/metadata"
	"Pando/test/mock"
	"context"
	bsrv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	dag "github.com/ipfs/go-merkledag"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

var (
	testCid1, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ = testCid1.Prefix().Sum([]byte("testdata2"))
	testCid3, _ = testCid1.Prefix().Sum([]byte("testdata3"))
)

func TestReceiveRecordAndOutUpdate_(t *testing.T) {
	Convey("test metadata manager", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		mm, err := New(context.Background(), pando.DS, pando.BS)
		So(err, ShouldBeNil)
		mockRecord := []*MetaRecord{
			{testCid1, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV6", uint64(time.Now().UnixNano())},
			{testCid2, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4", uint64(time.Now().UnixNano())},
			{testCid3, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV5", uint64(time.Now().UnixNano())},
		}

		Convey("test receive record and out update", func() {
			recvCh := mm.GetMetaInCh()
			for _, r := range mockRecord {
				recvCh <- r
			}
			outCh := mm.GetUpdateOut()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

			t.Cleanup(func() {
				cancel()
			})

			select {
			case <-ctx.Done():
				t.Error("timeout!not get update rightly")
			case update := <-outCh:
				So(len(update), ShouldEqual, 3)
				So(update, ShouldContainKey, mockRecord[0].ProviderID)
				So(update, ShouldContainKey, mockRecord[1].ProviderID)
				So(update, ShouldContainKey, mockRecord[2].ProviderID)
				So(update[mockRecord[0].ProviderID].Cidlist, ShouldResemble, []cid.Cid{testCid1})
				So(update[mockRecord[1].ProviderID].Cidlist, ShouldResemble, []cid.Cid{testCid2})
				So(update[mockRecord[2].ProviderID].Cidlist, ShouldResemble, []cid.Cid{testCid3})
			}
		})
		Convey("test export car", func() {
			dags := dag.NewDAGService(bsrv.New(pando.BS, offline.Exchange(pando.BS)))
			So(dags, ShouldNotBeNil)
			provider, err := mock.NewMockProvider(pando)
			So(err, ShouldBeNil)
			err = pando.Core.Subscribe(context.Background(), provider.ID)
			So(err, ShouldBeNil)
			daglist, err := provider.SendDag()
			So(err, ShouldBeNil)
			time.Sleep(time.Second * 5)
			tmpdir := t.TempDir()
			carpath := tmpdir + time.Now().String() + ".car"
			err = ExportMetaCar(dags, []cid.Cid{daglist[0]}, carpath, pando.BS)
			So(err, ShouldBeNil)
			data, err := os.ReadFile(carpath)
			So(err, ShouldBeNil)
			So(data, ShouldNotBeNil)
		})

	})
}
