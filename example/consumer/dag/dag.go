package main

import (
	"fmt"
	peerHelper "github.com/kenlabs/pando/pkg/util/peer"
	consumerSdk "github.com/kenlabs/pando/sdk/pkg/consumer"
	"time"
)

const (
	privateKeyStr  = "CAESQAycIStrQXBoxgf2pEazDLoZbL8WCLX5GIb69dl4x2mJMpukCAPbzq1URPtKen4Bpxfz9et2exWhfAfZ/RG30ts="
	pandoAddr      = "/ip4/52.14.211.248/tcp/9013"
	pandoPeerID    = "12D3KooWNU48MUrPEoYh77k99RbskgftfmSm3CdkonijcM5VehS9"
	providerPeerID = "12D3KooWNnK4gnNKmh6JUzRb34RqNcBahN5B8v18DsMxQ8mCqw81"
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

	consumer, err := consumerSdk.NewDAGConsumer(privateKeyStr, "http://52.14.211.248:9011", connectTimeout, syncTimeout)
	if err != nil {
		panic(err)
	}

	err = consumer.Start(pandoAddr, pandoPeerID, providerPeerID)
	if err != nil {
		panic(err)
	}
}
