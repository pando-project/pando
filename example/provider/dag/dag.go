package main

import (
	"fmt"
	peerHelper "github.com/kenlabs/pando/pkg/util/peer"
	"os"
	"os/signal"
	"syscall"
	"time"

	pandoSdk "github.com/kenlabs/pando/sdk/pkg/provider"
)

const (
	privateKeyStr = "CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w="
	pandoAddr     = "/ip4/127.0.0.1/tcp/9002"
	pandoPeerID   = "12D3KooWQKBbF3UkqzmJhh14u8FzZzFJMLCd5SjYbeM8Gk8HTrjz"
)

func main() {
	peerID, err := peerHelper.GetPeerIDFromPrivateKeyStr(privateKeyStr)
	if err != nil {
		panic(err)
	}
	fmt.Printf("provider peerID: %v\n", peerID.String())

	//dags := generateDAGs(1)
	provider, err := pandoSdk.NewDAGProvider(privateKeyStr, 10*time.Second, 10*time.Minute)
	if err != nil {
		panic(err)
	}

	err = provider.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		panic(err)
	}

	fmt.Println("pushing data to Pando...")
	metadata1, err := provider.NewMetadata([]byte("doge13234"))
	if err != nil {
		panic(err)
	}
	metadata1Cid, err := provider.Push(metadata1)
	if err != nil {
		panic(err)
	}
	metadata2, err := provider.AppendMetadata(metadata1, []byte("kitty13234"))
	if err != nil {
		panic(err)
	}
	metadata2Cid, err := provider.Push(metadata2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("pushed 2 nodes: \n\t%s\n\t%s\n", metadata1Cid.String(), metadata2Cid.String())

	//time.Sleep(20 * time.Second)

	// test for redundant push
	//_, _ = provider.Push(metadata1)

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
