package statetree

import (
	"context"
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	types2 "pando/pkg/statetree/types"
	"pando/test/mock"
	"reflect"
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

		ex := &types2.ExtraInfo{}
		ch := make(chan map[peer.ID]*types2.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, ex)
		So(err, ShouldBeNil)

		mockUpdate := map[peer.ID]*types2.ProviderState{
			testPeer: {
				Cidlist: []cid.Cid{testCid1, testCid2, testCid3},
			},
		}

		ch <- mockUpdate
		time.Sleep(time.Millisecond * 500)

		ss, err := st.GetSnapShotByHeight(0)
		So(ss.Height, ShouldEqual, uint64(0))
		So(st.height, ShouldEqual, uint64(1))
		ss, err = st.GetSnapShotByHeight(1000)
		So(err, ShouldResemble, NotFoundErr)
		//ss, err = st.GetSnapShotByHeight(-1)
		//So(err, ShouldResemble, fmt.Errorf("height must be positive"))

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
	Convey("when call api then get pando info", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		So(core, ShouldNotBeNil)

		ex := &types2.ExtraInfo{}
		ch := make(chan map[peer.ID]*types2.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, ex)
		So(err, ShouldBeNil)
		_ex, err := st.GetPandoInfo()
		So(_ex, ShouldResemble, &types2.ExtraInfo{})
		err = st.Shutdown()
		So(err, ShouldBeNil)
		st, err = New(context.Background(), core.DS, core.BS, ch, nil)
		_ex, err = st.GetPandoInfo()
		So(err, ShouldResemble, fmt.Errorf("nil info"))
		So(_ex, ShouldBeNil)
	})
}

func TestStateTreeDeleteDS(t *testing.T) {
	Convey("when delete statetree data then get nil in ds", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		So(core, ShouldNotBeNil)

		ex := &types2.ExtraInfo{}
		ch := make(chan map[peer.ID]*types2.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, ex)
		So(err, ShouldBeNil)

		ctx := context.Background()

		err = st.DeleteInfo(ctx)
		So(err, ShouldBeNil)
		root, err := pando.DS.Get(ctx, datastore.NewKey(RootKey))
		So(err, ShouldResemble, datastore.ErrNotFound)
		So(root, ShouldBeNil)
		snapShotList, err := pando.DS.Get(ctx, datastore.NewKey(SnapShotList))
		So(err, ShouldResemble, datastore.ErrNotFound)
		So(snapShotList, ShouldBeNil)

	})
}

func TestStateTreeInitFailed(t *testing.T) {
	Convey("when failed to init then reinitialize or return error", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		So(core, ShouldNotBeNil)
		ctx := context.Background()
		err = pando.DS.Put(ctx, datastore.NewKey(RootKey), []byte("testdata"))
		So(err, ShouldBeNil)

		ch := make(chan map[peer.ID]*types2.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, nil)
		So(err, ShouldResemble, fmt.Errorf("failed to load the State root from datastore"))

		err = pando.DS.Put(ctx, datastore.NewKey(RootKey), testCid1.Bytes())
		st, err = New(context.Background(), core.DS, core.BS, ch, nil)
		So(err, ShouldBeNil)
		err = st.Shutdown()
		So(err, ShouldBeNil)
	})

}

func TestStateTreeErrorHandle(t *testing.T) {
	Convey("when got unexpected errors then handle them", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		core := pando.Core
		So(core, ShouldNotBeNil)
		patch := gomonkey.ApplyMethod(reflect.TypeOf(&StateTree{}), "UpdateRoot", func(_ *StateTree, _ context.Context, _ map[peer.ID]*types2.ProviderState) (cid.Cid, error) {
			return cid.Undef, fmt.Errorf("unknown error")
		})
		ch := make(chan map[peer.ID]*types2.ProviderState)
		st, err := New(context.Background(), core.DS, core.BS, ch, nil)
		So(err, ShouldBeNil)
		mockUpdate := map[peer.ID]*types2.ProviderState{
			testPeer: {
				Cidlist: []cid.Cid{testCid1, testCid2, testCid3},
			},
		}
		ch <- mockUpdate
		time.Sleep(time.Millisecond * 500)
		err = st.Shutdown()
		So(err, ShouldBeNil)

		patch.Reset()
		patch = gomonkey.ApplyMethod(reflect.TypeOf(st), "CreateSnapShot", func(_ *StateTree, _ context.Context, newRoot cid.Cid, update map[peer.ID]*types2.ProviderState) error {
			return fmt.Errorf("unknown error")
		})
		defer patch.Reset()
		ch = make(chan map[peer.ID]*types2.ProviderState)
		st, err = New(context.Background(), core.DS, core.BS, ch, nil)
		So(err, ShouldBeNil)
		mockUpdate = map[peer.ID]*types2.ProviderState{
			testPeer: {
				Cidlist: []cid.Cid{testCid1, testCid2, testCid3},
			},
		}
		ch <- mockUpdate
		time.Sleep(time.Millisecond * 500)
		err = st.Shutdown()
		So(err, ShouldBeNil)

	})
}
