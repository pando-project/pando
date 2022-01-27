package main

import (
	"fmt"
	"os"
	"os/signal"
	peerHelper "pando/pkg/util/peer"
	"syscall"
	"time"

	pandoSdk "pando/sdk/pkg/provider"
	schema2 "pando/types/schema"
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

	provider, linkSystem, err := pandoSdk.NewDAGProvider(privateKeyStr, 10*time.Second, 10*time.Minute)
	if err != nil {
		panic(err)
	}

	metadata1, _ := schema2.NewMetadata(nil)
	metadataLink1, err := schema2.MetadataLink(linkSystem, metadata1)
	if err != nil {
		panic(err)
	}

	metadata2, _ := schema2.NewMetadata(metadataLink1)
	metadataLink2, err := schema2.MetadataLink(linkSystem, metadata2)
	if err != nil {
		panic(err)
	}

	metadata3, _ := schema2.NewMetadata(metadataLink2)
	metadataLink3, err := schema2.MetadataLink(linkSystem, metadata3)
	if err != nil {
		panic(err)
	}

	err = provider.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		panic(err)
	}

	fmt.Println("publishing data to Pando...")
	provider.Publish(metadataLink3.ToCid())

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
