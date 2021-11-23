package metadata

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/stretchr/testify/assert"
	//"gotest.tools/assert"
	"testing"
	"time"
)

var (
	testCid1, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ = testCid1.Prefix().Sum([]byte("testdata2"))
	testCid3, _ = testCid1.Prefix().Sum([]byte("testdata3"))
)

func TestCreate(t *testing.T) {
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)

	_, err := New(context.Background(), mds, bs)
	assert.NoError(t, err)
}

func TestReceiveRecordAndOutUpdate(t *testing.T) {
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)

	mm, err := New(context.Background(), mds, bs)
	if err != nil {
		t.Error(err)
	}

	t.Log(testCid1.String())
	t.Log(testCid2.String())
	t.Log(testCid3.String())

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	t.Cleanup(func() {
		cancel()
	})

	select {
	case <-ctx.Done():
		t.Error("timeout!not get update rightly")
	case update := <-outCh:
		assert.Equal(t, len(update), 3)
		assert.Contains(t, update, mockRecord[0].ProviderID)
		assert.Contains(t, update, mockRecord[1].ProviderID)
		assert.Contains(t, update, mockRecord[2].ProviderID)
		assert.Equal(t, update[mockRecord[0].ProviderID].Cidlist, []cid.Cid{testCid1})
		assert.Equal(t, update[mockRecord[1].ProviderID].Cidlist, []cid.Cid{testCid2})
		assert.Equal(t, update[mockRecord[2].ProviderID].Cidlist, []cid.Cid{testCid3})
	}
}
