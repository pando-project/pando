package main

import (
	"Pando/legs"
	"Pando/mock_provider/task"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime/datamodel"
	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipld/go-ipld-prime"

	golegs "github.com/filecoin-project/go-legs"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multicodec"
)

func Store(bs blockstore.Blockstore, n ipld.Node) (ipld.Link, error) {
	linkproto := cidlink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagJson),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: 16,
		},
	}
	lsys := legs.MkLinkSystem(bs)

	return lsys.Store(ipld.LinkContext{}, linkproto, n)
}

// eg: CAESQHWlReUYxW7FDvTAAqG+kNH2U7khW+iv0r+070+zKmFn9t80v5e30/NsBx5XzBLCE4uH/h3d3tpXlwCuO4YGN+w= 10 12D3KooWC3jxxw4TdQtoZDv3QNwmh9rtuiyVL8CADpnJYKHh9AiA /ip4/52.14.211.248/tcp/9000 30
func main() {
	if len(os.Args) < 5 {
		fmt.Println("please input:\r\n1. provider private key\n2. mock block number\n3. Pando PeerID\n4. Pando MultiAddr\n5. Time wait for data transferring[optional, int]")
		os.Exit(1)
	}

	privstr := os.Args[1]
	blocknum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("block number must be integer, not: ", os.Args[2])
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

	//srcStore := dssync.MutexWrap(datastore.NewMapDatastore())
	srcStore := dssync.MutexWrap(dstore)
	h, _ := libp2p.New(context.Background(),
		libp2p.Identity(privkey),
	)
	fmt.Println("p2pHost addr:", h.Addrs())
	fmt.Println("p2pHost id:", h.ID())
	bs := blockstore.NewBlockstore(srcStore)
	srcLnkS := legs.MkLinkSystem(bs)
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

	// gen test nodes
	nodes := make([]datamodel.Node, 0)
	mockTasks := task.GenMockTask(blocknum)
	for i := 0; i < len(mockTasks); i++ {
		taskBytes, err := json.Marshal(mockTasks[i])
		if err != nil {
			log.Fatal(err)
		}

		nn := basicnode.NewString(string(taskBytes))
		nodes = append(nodes, nn)
		fmt.Println(string(taskBytes))
	}

	lp, err := golegs.NewPublisher(context.Background(), h, srcStore, srcLnkS, "PandoPubSub")
	if err != nil {
		log.Fatal(err)
	}
	defer lp.Close()
	time.Sleep(time.Second * 2)
	lnks := make([]ipld.Link, 0)
	for i := 0; i < len(mockTasks); i++ {
		lk, err := Store(bs, nodes[i])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Add cid %s to root\r\n", lk.String())
		lnks = append(lnks, lk)
	}

	time.Sleep(time.Second)
	for i := 0; i < len(mockTasks); i++ {
		err = lp.UpdateRoot(context.Background(), lnks[i].(cidlink.Link).Cid)
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
