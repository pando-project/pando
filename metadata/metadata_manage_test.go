package metadata

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"testing"
	"time"
)

var (
	testCid1, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfab")
	testCid3, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfac")
)

func TestCreate(t *testing.T) {
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)

	_, err := New(context.Background(), mds, bs)
	if err != nil {
		t.Error(err)
	}
}

func TestReceiveRecordAndOutUpadte(t *testing.T) {
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)

	mm, err := New(context.Background(), mds, bs)
	if err != nil {
		t.Error(err)
	}

	mockRecord := []*MetaRecord{
		{testCid1, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV6", uint64(time.Now().UnixNano())},
		{testCid2, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4", uint64(time.Now().UnixNano())},
		{testCid3, "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV5", uint64(time.Now().UnixNano())},
	}

	recvCh := mm.GetMetaInCh()
	for _, r := range mockRecord {
		recvCh <- r
	}
	outCh := mm.GetUpdateOut()
	time.Sleep(time.Second * 6)
	select {
	case update := <-outCh:
		t.Log(update)
	default:
		t.Error("not get update rightly")
	}
}
