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
	"github.com/stretchr/testify/assert"
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
	testCid2, _ = testCid1.Prefix().Sum([]byte("testdata2"))
	testCid3, _ = testCid1.Prefix().Sum([]byte("testdata3"))
	testPeer, _ = peer.Decode("12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4")
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
	assert.NoError(t, err)
}

func TestStateTreeRoundTrip(t *testing.T) {
	core := getMockCore()
	ex := &types.ExtraInfo{
		GraphSyncUrl:   "",
		GoLegsSubUrl:   "",
		GolegsSubTopic: "",
		MultiAddr:      "test",
	}
	ch := make(chan map[peer.ID]*types.ProviderState)
	st, err := New(context.Background(), core.DS, core.BS, ch, ex)
	assert.NoError(t, err)

	mockUpdate := map[peer.ID]*types.ProviderState{
		testPeer: {
			Cidlist: []cid.Cid{testCid1, testCid2, testCid3},
		},
	}

	ch <- mockUpdate

	time.Sleep(time.Millisecond * 500)

	ss, err := st.GetSnapShotByHeight(0)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, ss.Height, uint64(0))
	assert.Equal(t, st.height, uint64(1))

	pstate, err := st.GetProviderStateByPeerID(testPeer)
	assert.NoError(t, err)
	l, err := st.GetSnapShotCidList()
	assert.NoError(t, err)
	if len(l) < 1 {
		t.Error("wrong snapshot cidlist ", l)
	}

	ch <- mockUpdate
	time.Sleep(time.Millisecond * 800)
	l, err = st.GetSnapShotCidList()
	assert.NoError(t, err)

	if len(l) < 2 {
		t.Fatal("wrong snapshot cidlist ", l)
	}

	_, err = st.GetSnapShot(l[0])
	assert.NoError(t, err)

	_, err = st.GetSnapShot(l[1])
	assert.NoError(t, err)

	h := st.height
	assert.Equal(t, h, uint64(2), "wrong height")

	err = st.Shutdown()
	assert.NoError(t, err)

	st, err = New(context.Background(), core.DS, core.BS, ch, ex)
	assert.NoError(t, err)

	assert.Equal(t, st.height, uint64(2))

	pstate, err = st.GetProviderStateByPeerID(testPeer)
	assert.NoError(t, err)

	assert.Equal(t, pstate.State.LastCommitHeight, uint64(1))
	t.Log(pstate.NewestUpdate)

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
	assert.NoError(t, err)

	close(ch)

	time.Sleep(time.Second)
}
