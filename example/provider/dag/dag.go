package main

import (
	"fmt"
	"os"
	"os/signal"
	peerHelper "pando/pkg/util/peer"
	"syscall"
	"time"

	"github.com/hashicorp/go-uuid"
	ipldFormat "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"

	pandoSdk "pando/sdk/pkg/provider"
)

const (
	privateKeyStr = "CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w="
	pandoAddr     = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID   = "12D3KooWMQ2gFA58MgxKtxr7xvjvg7dPPMqrEULGgVrrrD3Fa8cB"
)

func main() {
	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("provider peerID: %v\n", peerID.String())

	dags := generateDAGs(1)
	provider, err := pandoSdk.NewDAGProvider(privateKeyStr, 10*time.Second, 10*time.Minute)
	if err != nil {
		panic(err)
	}

	err = provider.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		panic(err)
	}

	fmt.Println("pushing data to Pando...")
	for _, dag := range dags {
		err := provider.PushMany(dag)
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(20 * time.Second)
	fmt.Println("press ctrl+c to exit.")

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down provider...")
	err = provider.Close()
	if err != nil {
		panic(err)
	}
	fmt.Println("Bye!")
}

func generateDAGs(quantity int) (dags [][]ipldFormat.Node) {
	for i := 0; i < quantity; i++ {
		kittyID := merkledag.NewRawNode([]byte(randID()))
		dogeID := merkledag.NewRawNode([]byte(randID()))
		piggyID := merkledag.NewRawNode([]byte(randID()))

		kittyNode := &merkledag.ProtoNode{}
		err := kittyNode.AddNodeLink("kittyID", kittyID)
		if err != nil {
			return nil
		}

		dogeNode := &merkledag.ProtoNode{}
		err = dogeNode.AddNodeLink("kittyNode", kittyNode)
		if err != nil {
			return nil
		}
		err = dogeNode.AddNodeLink("dogeID", dogeID)
		if err != nil {
			return nil
		}

		piggyNode := &merkledag.ProtoNode{}
		err = piggyNode.AddNodeLink("dogeNode", dogeNode)
		if err != nil {
			return nil
		}
		err = piggyNode.AddNodeLink("piggyID", piggyID)
		if err != nil {
			return nil
		}

		dags = append(dags, []ipldFormat.Node{
			piggyNode, dogeNode, kittyNode,
			piggyID, dogeID, kittyID,
		})
	}

	return dags
}

func randID() string {
	id, _ := uuid.GenerateUUID()

	return id
}
