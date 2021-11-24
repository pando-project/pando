package main

import (
	"Pando/legs"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/ipfs/go-blockservice"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	golegs "github.com/filecoin-project/go-legs"
	dssync "github.com/ipfs/go-datastore/sync"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/libp2p/go-libp2p"
)

func getDagNodes(n int) [][]format.Node {
	var dags [][]format.Node
	for i := 0; i < n; i++ {
		a := merkledag.NewRawNode([]byte("aaaaa" + string(rune(rand.Intn(100000)))))
		b := merkledag.NewRawNode([]byte("bbbbb" + string(rune(rand.Intn(100000)))))
		c := merkledag.NewRawNode([]byte("ccccc" + string(rune(rand.Intn(100000)))))

		nd1 := &merkledag.ProtoNode{}
		nd1.AddNodeLink("cat", a)

		nd2 := &merkledag.ProtoNode{}
		nd2.AddNodeLink("first", nd1)
		nd2.AddNodeLink("dog", b)

		nd3 := &merkledag.ProtoNode{}
		nd3.AddNodeLink("second", nd2)
		nd3.AddNodeLink("bear", c)
		dags = append(dags, []format.Node{nd3, nd2, nd1, c, b, a})
	}

	return dags
}

// eg: CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w= 1 12D3KooWC3jxxw4TdQtoZDv3QNwmh9rtuiyVL8CADpnJYKHh9AiA /ip4/52.14.211.248/tcp/9000 30
func main() {
	if len(os.Args) < 5 {
		fmt.Println("please input:\r\n1. provider private key\n2. mock block number\n3. Pando PeerID\n4. Pando MultiAddr\n5. Time wait for data transferring[optional, int]")
		os.Exit(1)
	}

	privstr := os.Args[1]
	dagnum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("dag number must be integer, not: ", os.Args[2])
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
	fmt.Println("p2pHost addr:", h.Addrs())
	fmt.Println("p2pHost id:", h.ID())
	bs := blockstore.NewBlockstore(srcStore)
	srcLnkS := legs.MkLinkSystem(bs)
	dags := merkledag.NewDAGService(blockservice.New(bs, offline.Exchange(bs)))
	// connect Pando
	ma, err := multiaddr.NewMultiaddr(PandoAddrStr + "/ipfs/" + PandoPeerID)
	if err != nil {
		log.Fatal(err)
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		log.Fatal(err)
	}
	if err = h.Connect(context.Background(), *peerInfo); err != nil {
		log.Fatal(err)
	}
	// create provider legs
	lp, err := golegs.NewPublisher(context.Background(), h, srcStore, srcLnkS, "PandoPubSub")
	if err != nil {
		log.Fatal(err)
	}
	defer lp.Close()
	time.Sleep(time.Second * 2)

	// store test dag
	dagsNodes := getDagNodes(dagnum)
	for _, dagNodes := range dagsNodes {
		for i := 0; i < len(dagNodes); i++ {
			err = dags.Add(context.Background(), dagNodes[i])
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	time.Sleep(time.Second)
	for _, dagNodes := range dagsNodes {
		fmt.Printf("the dag root cid is: %s\r\n", dagNodes[0].Cid())
		err = lp.UpdateRoot(context.Background(), dagNodes[0].Cid())
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("waiting for data transferring completed")

	if timesleep > 0 {
		time.Sleep(time.Second * time.Duration(timesleep))
	} else {
		time.Sleep(time.Second * 30)
	}
	return
}
