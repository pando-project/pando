package main

import (
	"fmt"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	peerHelper "github.com/pando-project/pando/pkg/util/peer"
	consumerSdk "github.com/pando-project/pando/sdk/pkg/consumer"
	"time"
)

const (
	privateKeyStr  = "CAESQAycIStrQXBoxgf2pEazDLoZbL8WCLX5GIb69dl4x2mJMpukCAPbzq1URPtKen4Bpxfz9et2exWhfAfZ/RG30ts="
	pandoAddr      = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID    = "12D3KooWNU48MUrPEoYh77k99RbskgftfmSm3CdkonijcM5VehS9"
	providerPeerID = "12D3KooWNnK4gnNKmh6JUzRb34RqNcBahN5B8v18DsMxQ8mCqw81"
)

const (
	connectTimeout = time.Minute
	syncTimeout    = 10 * time.Minute
	syncDepth      = 0
)

func main() {
	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("consumer peerID: %v\n", peerID.String())

	ch := make(chan Status, 0)

	// compute for success rate got from linksystem
	go func() {
		var total float32
		var success float32
		for status := range ch {
			total += 1
			if status == Status(4) {
				success += 1
			}
			successRate := success / total
			fmt.Printf("total job numbers: %d\n", int(total))
			fmt.Println("success rate: ", successRate)
		}
	}()

	// init store (mem for example), use custom link system for calculating
	ds := datastore.NewMapDatastore()
	mds := sync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	lsys := MkLinkSystem(bs, ch)

	consumer, err := consumerSdk.NewDAGConsumer(privateKeyStr, "http://52.14.211.248:9011", connectTimeout, &lsys, syncTimeout)
	if err != nil {
		panic(err)
	}

	var sel ipld.Node
	if syncDepth == 0 {
		sel = nil
	} else {
		sel = golegs.ExploreRecursiveWithStopNode(selector.RecursionLimitDepth(syncDepth), nil, nil)
	}

	err = consumer.Start(pandoAddr, pandoPeerID, providerPeerID, sel)
	if err != nil {
		panic(err)
	}
}
