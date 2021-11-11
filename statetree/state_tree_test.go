package statetree

import (
	"Pando/statetree/types"
	"context"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"testing"
	"time"
)

type MockCore struct {
	DS   *dssync.MutexDatastore
	BS   blockstore.Blockstore
	Host host.Host
}

var (
	testCid1, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfab")
	testCid3, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfac")
)

func getMockCore() *MockCore {
	ds := dssync.MutexWrap(datastore.NewMapDatastore())
	bs := blockstore.NewBlockstore(ds)
	h, _ := libp2p.New(context.Background())
	return &MockCore{
		DS:   ds,
		BS:   bs,
		Host: h,
	}
}

func TestNew(t *testing.T) {
	core := getMockCore()
	ex := &types.ExtraInfo{
		GraphSyncUrl:   "",
		GoLegsSubUrl:   "",
		GolegsSubTopic: "",
		MultiAddr:      "test",
	}
	ch := make(<-chan map[peer.ID]*types.ProviderState)
	_, err := New(context.Background(), core.DS, core.BS, ch, ex)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestStateTree_Update(t *testing.T) {
	core := getMockCore()
	ex := &types.ExtraInfo{
		GraphSyncUrl:   "",
		GoLegsSubUrl:   "",
		GolegsSubTopic: "",
		MultiAddr:      "test",
	}
	ch := make(chan map[peer.ID]*types.ProviderState)
	st, err := New(context.Background(), core.DS, core.BS, ch, ex)
	if err != nil {
		t.Error(err.Error())
	}

	mockUpdate := map[peer.ID]*types.ProviderState{
		"12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4": {
			[]cid.Cid{testCid1, testCid2, testCid3},
		},
	}

	ch <- mockUpdate

	time.Sleep(time.Millisecond * 500)

	_, err = st.GetSnapShotByHeight(0)
	if err != nil {
		t.Error(err)
	}

	l, err := st.GetSnapShotCidList()
	if err != nil {
		t.Error(err)
	}
	if len(l) < 1 {
		t.Error("wrong snapshot cidlist ", l)
	}

	ch <- mockUpdate
	time.Sleep(time.Millisecond * 500)
	l, err = st.GetSnapShotCidList()
	if err != nil {
		t.Error(err)
	}
	if len(l) < 2 {
		t.Error("wrong snapshot cidlist ", l)
	}

	_, err = st.GetSnapShot(l[0])
	if err != nil {
		t.Error(err)
	}
	_, err = st.GetSnapShot(l[1])
	if err != nil {
		t.Error(err)
	}

	err = st.Shutdown()
	if err != nil {
		t.Error(err)
	}
	st, err = New(context.Background(), core.DS, core.BS, ch, ex)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestCloseUpdateCh(t *testing.T) {
	core := getMockCore()
	ex := &types.ExtraInfo{
		GraphSyncUrl:   "",
		GoLegsSubUrl:   "",
		GolegsSubTopic: "",
		MultiAddr:      "test",
	}
	ch := make(chan map[peer.ID]*types.ProviderState)
	_, err := New(context.Background(), core.DS, core.BS, ch, ex)
	if err != nil {
		t.Error(err.Error())
	}
	close(ch)

	time.Sleep(time.Second)
}
