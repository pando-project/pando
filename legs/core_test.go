package legs

import (
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"Pando/metadata"
	"Pando/policy"
	"context"
	"fmt"
	goLegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipld/go-ipld-prime"
	cidLink "github.com/ipld/go-ipld-prime/linking/cid"
	basicNode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multicodec"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

var (
	tokenRate  = math.Ceil((0.8 * float64(1)) * 1)
	rateConfig = &policy.LimiterConfig{
		TotalRate:     tokenRate,
		TotalBurst:    int(math.Ceil(tokenRate)),
		BaseTokenRate: tokenRate,
		Registry:      newRegistry(),
	}
)

func Store(n ipld.Node, linkSystem ipld.LinkSystem) (ipld.Link, error) {
	linkPrototype := cidLink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagJson),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: 16,
		},
	}

	return linkSystem.Store(ipld.LinkContext{}, linkPrototype, n)
}

func newRegistry() *registry.Registry {
	dstore, err := leveldb.NewDatastore("/tmp", nil)
	lotusDiscoverer, err := lotus.NewDiscoverer("https://api.chain.love")
	registryInstance, err := registry.NewRegistry(
		&config.Discovery{Policy: config.Policy{Allow: true}},
		&config.AccountLevel{Threshold: []int{1, 10}}, dstore, lotusDiscoverer)
	if err != nil {
		panic(fmt.Errorf("new registry failed, error: %v", err))
	}
	return registryInstance
}

func getDagNodes() []format.Node {
	a := merkledag.NewRawNode([]byte("aaaaa"))
	b := merkledag.NewRawNode([]byte("bbbb"))
	c := merkledag.NewRawNode([]byte("cccc"))

	nd1 := &merkledag.ProtoNode{}
	err := nd1.AddNodeLink("cat", a)
	if err != nil {
		return nil
	}

	nd2 := &merkledag.ProtoNode{}
	if err = nd2.AddNodeLink("first", nd1); err != nil {
		return nil
	}
	if err := nd2.AddNodeLink("dog", b); err != nil {
		return nil
	}

	nd3 := &merkledag.ProtoNode{}
	if err := nd3.AddNodeLink("second", nd2); err != nil {
		return nil
	}
	if err := nd3.AddNodeLink("bear", c); err != nil {
		return nil
	}

	return []format.Node{nd3, nd2, nd1, c, b, a}
}

func TestCreate(t *testing.T) {
	host, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	ds := datastore.NewMapDatastore()
	mds := datastoreSync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	outCh := make(chan<- *metadata.MetaRecord)
	limiter, err := policy.NewLimiter(policy.LimiterConfig{
		TotalRate:  1,
		TotalBurst: 1,
	})
	if err != nil {
		t.Error(err)
	}
	_, err = NewLegsCore(context.Background(), &host, ds, bs, outCh, limiter)
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
	mds := datastoreSync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	outCh := make(chan *metadata.MetaRecord)
	limiter, err := policy.NewLimiter(*rateConfig)
	if err != nil {
		t.Error(err)
	}
	legsCore, err := NewLegsCore(context.Background(), &host, mds, bs, outCh, limiter)
	if err != nil {
		t.Error(err)
	}

	// mock provider legs
	srcHost, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	srcDatastore := datastoreSync.MutexWrap(datastore.NewMapDatastore())
	srcBlockstore := blockstore.NewBlockstore(srcDatastore)
	srcLinkSystem := MkLinkSystem(srcBlockstore)
	dags := merkledag.NewDAGService(blockservice.New(srcBlockstore, offline.Exchange(srcBlockstore)))
	legsPublisher, err := goLegs.NewPublisher(context.Background(), srcHost, srcDatastore, srcLinkSystem, "PandoPubSub")

	multiAddress := host.Addrs()[0].String() + "/ipfs/" + host.ID().String()
	peerInfo, err := peer.AddrInfoFromString(multiAddress)
	if err != nil {
		t.Fatal(err)
	}

	if err = srcHost.Connect(context.Background(), *peerInfo); err != nil {
		t.Fatal(err)
	}

	// mock Core subscribes the mock provider
	err = legsCore.Subscribe(context.Background(), srcHost.ID())
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

	err = legsPublisher.UpdateRoot(context.Background(), dagNodes[0].Cid())
	if err != nil {
		t.Fatal(err)
	}

	node := basicNode.NewString("test1")
	nlink, err := Store(node, srcLinkSystem)
	if err != nil {
		t.Error(err)
	}
	err = legsPublisher.UpdateRoot(context.Background(), nlink.(cidLink.Link).Cid)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	t.Cleanup(func() {
		cancel()
		if err := legsPublisher.Close(); err != nil {
			t.Error(err)
		}
		if err := legsCore.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})

	select {
	case _ = <-ctx.Done():
		t.Fatal("timeout!not receive record rightly")
	case record := <-outCh:
		assert.Equal(t, record.Cid, dagNodes[0].Cid())
		assert.Equal(t, record.ProviderID, srcHost.ID())
		t.Log(record)
	}

	select {
	case _ = <-ctx.Done():
		t.Fatal("timeout!not receive record rightly")
	case record := <-outCh:
		assert.Equal(t, record.Cid, nlink.(cidLink.Link).Cid, "expected: ", nlink.(cidLink.Link).Cid.String(), " received:", record.Cid.String())
		assert.Equal(t, record.ProviderID, srcHost.ID())
		t.Log(record)
	}

}

func TestLegsSync(t *testing.T) {
	host, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	ds := datastore.NewMapDatastore()
	mds := datastoreSync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	dags := merkledag.NewDAGService(blockservice.New(bs, offline.Exchange(bs)))
	outCh := make(chan *metadata.MetaRecord)
	limiter, err := policy.NewLimiter(policy.LimiterConfig{
		TotalRate:  1,
		TotalBurst: 1,
	})
	if err != nil {
		t.Error(err)
	}
	_, err = NewLegsCore(context.Background(), &host, mds, bs, outCh, limiter)
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
	dstHost, err := libp2p.New(context.Background())
	if err != nil {
		t.Error(err)
	}
	dstDatastore := datastoreSync.MutexWrap(datastore.NewMapDatastore())
	dstBlockstore := blockstore.NewBlockstore(dstDatastore)
	srcLinkSystem := MkLinkSystem(dstBlockstore)
	dstDags := merkledag.NewDAGService(blockservice.New(dstBlockstore, offline.Exchange(dstBlockstore)))

	ls, err := goLegs.NewSubscriber(context.Background(), dstHost, dstDatastore, srcLinkSystem, "PandoPubSub", nil)
	multiAddress := host.Addrs()[0].String() + "/ipfs/" + host.ID().String()
	peerInfo, err := peer.AddrInfoFromString(multiAddress)
	if err != nil {
		t.Fatal(err)
	}

	if err = dstHost.Connect(context.Background(), *peerInfo); err != nil {
		t.Fatal(err)
	}

	_, _, err = ls.Sync(context.Background(), host.ID(), dagNodes[0].Cid(), goLegs.LegSelector(nil))
	if err != nil {
		t.Fatal(err)
	}

	// wait graph-sync to save the block in blockstore
	time.Sleep(time.Second)

	_, err = dstDags.Get(context.Background(), dagNodes[0].Cid())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(dagNodes); i++ {
		_, err := dstBlockstore.Get(dagNodes[i].Cid())
		assert.NoError(t, err)
	}

}
