package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"log"
	"math/rand"
	"os"
	"pando/pkg/legs"
	schema2 "pando/types/schema"
	"strconv"
	"time"

	leveldb "github.com/ipfs/go-ds-leveldb"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	golegs "github.com/filecoin-project/go-legs"
	dssync "github.com/ipfs/go-datastore/sync"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/libp2p/go-libp2p"
)

func createMetadata(linkSystem ipld.LinkSystem) cid.Cid {
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

	return metadataLink3.ToCid()
}

// eg: CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w= 1 12D3KooWC3jxxw4TdQtoZDv3QNwmh9rtuiyVL8CADpnJYKHh9AiA /ip4/52.14.211.248/tcp/9000 30
func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage:")
		fmt.Println("    dag [privkey] [countdag] [peerid] [multiaddr] [wait]")
		fmt.Println()
		fmt.Println("Arguments:")
		fmt.Println("    privkey:   provider private key")
		fmt.Println("    countdag:  count of dag to send")
		fmt.Println("    peerid:    pando peer id")
		fmt.Println("    multiaddr: pando multiaddr")
		fmt.Println("    wait:      seconds of waiting time for data transfer (optional)")
		os.Exit(1)
	}

	privstr := os.Args[1]
	_, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("The count of dag must be integer, not: ", os.Args[2])
	}
	PandoPeerID := os.Args[3]
	PandoAddrStr := os.Args[4]
	var timesleep int
	if len(os.Args) > 5 {
		timesleep, err = strconv.Atoi(os.Args[5])
		if err != nil {
			timesleep = -1
		}
	}

	pkb, err := base64.StdEncoding.DecodeString(privstr)
	if err != nil {
		log.Fatal(err)
	}
	privkey, err := ic.UnmarshalPrivateKey(pkb)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())
	dstore, err := leveldb.NewDatastore("", nil)
	srcStore := dssync.MutexWrap(dstore)
	h, _ := libp2p.New(context.Background(),
		libp2p.Identity(privkey),
	)
	fmt.Println("MultiAddr:", h.Addrs())
	fmt.Println("Peer id:", h.ID())
	// bs := blockstore.NewBlockstore(srcStore)
	linkSystem := legs.MkLinkSystem(dstore)
	metadataRootCid := createMetadata(linkSystem)

	// connect Pando
	multiAddr, err := multiaddr.NewMultiaddr(PandoAddrStr + "/ipfs/" + PandoPeerID)
	if err != nil {
		log.Fatal(err)
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		log.Fatal(err)
	}
	if err = h.Connect(context.Background(), *peerInfo); err != nil {
		log.Fatal(err)
	}
	// create provider legs
	lp, err := golegs.NewPublisher(context.Background(), h, srcStore, linkSystem, "PandoPubSub")
	if err != nil {
		log.Fatal(err)
	}
	defer lp.Close()
	time.Sleep(time.Second * 2)

	err = lp.UpdateRoot(context.Background(), metadataRootCid)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("waiting for data transferring completed")

	if timesleep > 0 {
		time.Sleep(time.Second * time.Duration(timesleep))
	} else {
		time.Sleep(time.Second * 30)
	}
	return
}
