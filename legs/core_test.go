package legs

import (
	"Pando/metadata"
	"Pando/policy"
	"context"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multicodec"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Store(n ipld.Node, lsys ipld.LinkSystem) (ipld.Link, error) {
	linkproto := cidlink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagJson),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: 16,
		},
	}

	return lsys.Store(ipld.LinkContext{}, linkproto, n)
}

func getDagNodes() []format.Node {
	a := merkledag.NewRawNode([]byte("aaaaa"))
	b := merkledag.NewRawNode([]byte("bbbb"))
	c := merkledag.NewRawNode([]byte("cccc"))

	nd1 := &merkledag.ProtoNode{}
	nd1.AddNodeLink("cat", a)

	nd2 := &merkledag.ProtoNode{}
	nd2.AddNodeLink("first", nd1)
	nd2.AddNodeLink("dog", b)

	nd3 := &merkledag.ProtoNode{}
	nd3.AddNodeLink("second", nd2)
	nd3.AddNodeLink("bear", c)

	return []format.Node{nd3, nd2, nd1, c, b, a}
}

func TestCreate(t *testing.T) {
	host, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	outCh := make(chan<- *metadata.MetaRecord)
	_, err = NewLegsCore(context.Background(), &host, ds, bs, outCh, policy.NewLimiter(policy.LimiterConfig{}))
	if err != nil {
		t.Error(err)
	}

}

func TestGetMetaRecord(t *testing.T) {
	// create Core
	host, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	outCh := make(chan *metadata.MetaRecord)
	core, err := NewLegsCore(context.Background(), &host, mds, bs, outCh, policy.NewLimiter(policy.LimiterConfig{}))
	if err != nil {
		t.Error(err)
	}

	// mock provider legs
	srchost, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	srcds := datastore.NewMapDatastore()
	srcmds := dssync.MutexWrap(srcds)
	srcbs := blockstore.NewBlockstore(srcmds)
	srcLnkS := MkLinkSystem(srcbs)
	dags := merkledag.NewDAGService(blockservice.New(srcbs, offline.Exchange(srcbs)))
	lp, err := golegs.NewPublisher(context.Background(), srchost, srcmds, srcLnkS, "PandoPubSub")

	mastr := host.Addrs()[0].String() + "/ipfs/" + host.ID().String()
	peerInfo, err := peer.AddrInfoFromString(mastr)
	if err != nil {
		t.Fatal(err)
	}

	if err = srchost.Connect(context.Background(), *peerInfo); err != nil {
		t.Fatal(err)
	}

	// mock Core subscribe the mock provider
	err = core.Subscribe(context.Background(), srchost.ID())
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	// store test dag
	dagNodes := getDagNodes()
	for i := 0; i < len(dagNodes); i++ {
		err = dags.Add(context.Background(), dagNodes[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	err = lp.UpdateRoot(context.Background(), dagNodes[0].Cid())
	if err != nil {
		t.Fatal(err)
	}

	node := basicnode.NewString("test1")
	nlink, err := Store(node, srcLnkS)
	if err != nil {
		t.Error(err)
	}
	err = lp.UpdateRoot(context.Background(), nlink.(cidlink.Link).Cid)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	t.Cleanup(func() {
		cancel()
		lp.Close()
		core.Close(context.Background())
	})

	select {
	case _ = <-ctx.Done():
		t.Fatal("timeout!not receive record rightly")
	case record := <-outCh:
		assert.Equal(t, record.Cid, dagNodes[0].Cid())
		assert.Equal(t, record.ProviderID, srchost.ID())
		t.Log(record)
	}

	select {
	case _ = <-ctx.Done():
		t.Fatal("timeout!not receive record rightly")
	case record := <-outCh:
		assert.Equal(t, record.Cid, nlink.(cidlink.Link).Cid, "expected: ", nlink.(cidlink.Link).Cid.String(), " received:", record.Cid.String())
		assert.Equal(t, record.ProviderID, srchost.ID())
		t.Log(record)
	}

}

func TestLegsSync(t *testing.T) {
	host, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	dags := merkledag.NewDAGService(blockservice.New(bs, offline.Exchange(bs)))
	outCh := make(chan *metadata.MetaRecord)
	_, err = NewLegsCore(context.Background(), &host, mds, bs, outCh, policy.NewLimiter(policy.LimiterConfig{}))
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)

	// store test dag
	dagNodes := getDagNodes()
	for i := 0; i < len(dagNodes); i++ {
		err = dags.Add(context.Background(), dagNodes[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	// mock provider legs
	dsthost, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	dstds := datastore.NewMapDatastore()
	dstmds := dssync.MutexWrap(dstds)
	dstbs := blockstore.NewBlockstore(dstmds)
	srcLnkS := MkLinkSystem(dstbs)
	dstdags := merkledag.NewDAGService(blockservice.New(dstbs, offline.Exchange(dstbs)))

	ls, err := golegs.NewSubscriber(context.Background(), dsthost, dstmds, srcLnkS, "PandoPubSub", nil)
	mastr := host.Addrs()[0].String() + "/ipfs/" + host.ID().String()
	peerInfo, err := peer.AddrInfoFromString(mastr)
	if err != nil {
		t.Fatal(err)
	}

	if err = dsthost.Connect(context.Background(), *peerInfo); err != nil {
		t.Fatal(err)
	}

	_, _, err = ls.Sync(context.Background(), host.ID(), dagNodes[0].Cid(), golegs.LegSelector(nil))
	if err != nil {
		t.Fatal(err)
	}

	// wait graphsync to save the block in blockstore
	time.Sleep(time.Second)

	_, err = dstdags.Get(context.Background(), dagNodes[0].Cid())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(dagNodes); i++ {
		_, err := dstbs.Get(dagNodes[i].Cid())
		assert.NoError(t, err)
	}

}
