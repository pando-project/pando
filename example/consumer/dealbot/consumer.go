package main

import (
	"context"
	"encoding/base64"
	"fmt"
	dt "github.com/filecoin-project/go-data-transfer/impl"
	dtnetwork "github.com/filecoin-project/go-data-transfer/network"
	gstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	gsimpl "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	peerHelper "github.com/kenlabs/pando/pkg/util/peer"
	"github.com/kenlabs/pando/test/mock"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	privateKeyStr  = "CAESQAycIStrQXBoxgf2pEazDLoZbL8WCLX5GIb69dl4x2mJMpukCAPbzq1URPtKen4Bpxfz9et2exWhfAfZ/RG30ts="
	pandoAddr      = "/ip4/52.14.211.248/tcp/9013"
	pandoPeerID    = "12D3KooWNU48MUrPEoYh77k99RbskgftfmSm3CdkonijcM5VehS9"
	providerPeerID = "12D3KooWNnK4gnNKmh6JUzRb34RqNcBahN5B8v18DsMxQ8mCqw81"
)

func main() {
	//logging.SetLogLevel("*", "debug")

	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("consumer peerID: %v\n", peerID.String())

	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		panic(err)
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		panic(err)
	}

	consumerHost, err := libp2p.New(libp2p.Identity(privateKey))
	if err != nil {
		panic(err)
	}

	ds := datastore.NewMapDatastore()
	mds := sync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	ch := make(chan Status, 0)

	go func() {
		var total float32
		var success float32
		for status := range ch {
			total += 1
			if status == Status(4) {
				success += 1
			}
			successRate := success / total
			fmt.Println("success rate: ", successRate)
		}
	}()

	lsys := MkLinkSystem(bs, ch)
	graphSyncNet := gsnet.NewFromLibp2pHost(consumerHost)
	dataTransferNet := dtnetwork.NewFromLibp2pHost(consumerHost)
	graphExchange := gsimpl.New(context.Background(), graphSyncNet, lsys)
	graphTransport := gstransport.NewTransport(consumerHost.ID(), graphExchange)

	dataManager, err := dt.NewDataTransfer(mds, dataTransferNet, graphTransport)
	if err != nil {
		panic(err)
	}
	err = dataManager.Start(context.Background())
	if err != nil {
		panic(err)
	}
	consumer, err := golegs.NewSubscriber(consumerHost,
		nil,
		lsys,
		mock.GetTopic(),
		nil,
		golegs.DtManager(dataManager, graphExchange),
	)
	if err != nil {
		panic(err)
	}
	multiAddress := pandoAddr + "/ipfs/" + pandoPeerID
	peerInfo, err := peer.AddrInfoFromString(multiAddress)
	fmt.Println("connecting Pando.......")
	err = consumerHost.Connect(context.Background(), *peerInfo)
	if err != nil {
		panic(err)
	}

	cc, err := cid.Decode("baguqeeqqtaayr376ossvblkakdhnwoyk34")
	if err != nil {
		panic(err)
	}
	fmt.Println("Syncing.......")

	ctx, cncl := context.WithTimeout(context.Background(), time.Second*5)
	defer cncl()
	sel := golegs.ExploreRecursiveWithStopNode(selector.RecursionLimitDepth(5), nil, nil)
	c, err := consumer.Sync(ctx, peerInfo.ID, cc, sel, nil)
	if err != nil {
		panic(err)
	}

	fmt.Print(c.String())
}
