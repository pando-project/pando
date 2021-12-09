package statetree

import (
	"Pando/statetree/types"
	"Pando/test/mock"
	"context"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var (
	testCid1, _ = cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ = testCid1.Prefix().Sum([]byte("testdata2"))
	testCid3, _ = testCid1.Prefix().Sum([]byte("testdata3"))
	testPeer, _ = peer.Decode("12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4")
)

func TestStateTreeRoundTrip_(t *testing.T) {
	Convey("test state tree round trip", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		So(core, ShouldNotBeNil)

		ex := &types.ExtraInfo{}
		ch := make(chan map[peer.ID]*types.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, ex)
		So(err, ShouldBeNil)

		mockUpdate := map[peer.ID]*types.ProviderState{
			testPeer: {
				Cidlist: []cid.Cid{testCid1, testCid2, testCid3},
			},
		}

		ch <- mockUpdate
		time.Sleep(time.Millisecond * 500)

		ss, err := st.GetSnapShotByHeight(0)
		So(ss.Height, ShouldEqual, uint64(0))
		So(st.height, ShouldEqual, uint64(1))

		pstate, err := st.GetProviderStateByPeerID(testPeer)
		So(err, ShouldBeNil)
		l, err := st.GetSnapShotCidList()
		So(err, ShouldBeNil)
		So(len(l), ShouldEqual, 1)

		ch <- mockUpdate
		time.Sleep(time.Millisecond * 800)

		l, err = st.GetSnapShotCidList()
		So(err, ShouldBeNil)
		So(len(l), ShouldEqual, 2)

		ss, err = st.GetSnapShot(l[0])
		So(err, ShouldBeNil)
		So(ss.Height, ShouldEqual, uint64(0))
		So(st.height, ShouldEqual, uint64(2))

		ss, err = st.GetSnapShot(l[1])
		So(ss.Height, ShouldEqual, uint64(1))

		err = st.Shutdown()
		So(err, ShouldBeNil)

		st, err = New(context.Background(), core.DS, core.BS, ch, ex)
		So(err, ShouldBeNil)

		pstate, err = st.GetProviderStateByPeerID(testPeer)
		So(err, ShouldBeNil)
		So(pstate.State.LastCommitHeight, ShouldEqual, uint64(1))

		close(ch)
		So(<-st.ctx.Done(), ShouldNotBeNil)
	})
}
