package main

import (
	"fmt"
	"github.com/ipfs/go-cid"
	peerHelper "github.com/kenlabs/pando/pkg/util/peer"
	consumerSdk "github.com/kenlabs/pando/sdk/pkg/consumer"
	"time"
)

const (
	privateKeyStr = "CAESQAycIStrQXBoxgf2pEazDLoZbL8WCLX5GIb69dl4x2mJMpukCAPbzq1URPtKen4Bpxfz9et2exWhfAfZ/RG30ts="
	pandoAddr     = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID   = "12D3KooWJjPMqp1eAN6DAvDXJQGivWBq85EqFP29VkteePBKgesa"
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

	consumer, err := consumerSdk.NewDAGConsumer(privateKeyStr, connectTimeout, syncTimeout)
	err = consumer.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		panic(err)
	}
	latestSyncCid, err := consumer.Sync(cid.Undef, nil)
	fmt.Println("cid: ", latestSyncCid)
	if err != nil {
		panic(err)
	}
	fmt.Printf("sync succeed")

}
