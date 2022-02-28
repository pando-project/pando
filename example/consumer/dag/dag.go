package main

import (
	"context"
	"fmt"
	peerHelper "github.com/kenlabs/pando/pkg/util/peer"
	consumerSdk "github.com/kenlabs/pando/sdk/pkg/consumer"
	"time"
)

const (
	privateKeyStr  = "CAESQAycIStrQXBoxgf2pEazDLoZbL8WCLX5GIb69dl4x2mJMpukCAPbzq1URPtKen4Bpxfz9et2exWhfAfZ/RG30ts="
	pandoAddr      = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID    = "12D3KooWJjPMqp1eAN6DAvDXJQGivWBq85EqFP29VkteePBKgesa"
	providerPeerID = "12D3KooWSS3sEujyAXB9SWUvVtQZmxH6vTi9NitqaaRQoUjeEk3M"
)

const (
	connectTimeout = time.Minute
	syncTimeout    = 10 * time.Minute
)

func main() {
	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("consumer peerID: %v\n", peerID.String())

	consumer, err := consumerSdk.NewDAGConsumer(privateKeyStr, "http://127.0.0.1:9000", connectTimeout, syncTimeout)
	err = consumer.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		panic(err)
	}
	headCid, err := consumer.GetLatestHead(providerPeerID)
	if err != nil {
		panic(err)
	}
	latestSyncCid, err := consumer.Sync(headCid, nil)
	fmt.Println("cid: ", latestSyncCid)
	if err != nil {
		panic(err)
	}
	fmt.Println("sync succeed")

	block, err := consumer.Core.Blockstore.Get(context.Background(), latestSyncCid)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", block.RawData())
}
