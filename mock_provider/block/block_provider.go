package main

import (
	"Pando/config"
	"Pando/legs"
	"Pando/mock_provider/task"
	"context"
	"encoding/json"
	"fmt"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime/datamodel"
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

var (
	//mockTasksNum         = 5
	TestProviderIdentity = &config.Identity{
		PeerID:  "12D3KooWDi135q9xcE7xiRN1bBZZGc15dSyFRgm7pajTmt7ndCX5",
		PrivKey: "CAESQHMFRinebmZ/C2zo8tJfYlWxrW5jUIaNoKndLO/LNuLlOc1eZZUi3InQk7QIx0ggEBtkisx7wd+bFsYJrjkc2Uw=",
	}
	PrivKey, _ = TestProviderIdentity.DecodePrivateKey("")
	// PandoAddrStr Pando Info
	//PandoAddrStr = "/ip4/192.168.0.101/tcp/5003"
	//PandoPeerID  = "12D3KooWCjMkPdoB9vWQwC2e98yBB9fK6t4mb9c7tZeDuMpSDmq3"
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

func main() {
	if len(os.Args) < 4 {
		fmt.Println("please input:\r\n 1.mockTask number 2. Pando PeerID 3. Pando MultiAddr")
		os.Exit(1)
	}

	mockTasksNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err.Error())
	}
	PandoPeerID := os.Args[2]
	PandoAddrStr := os.Args[3]

	rand.Seed(time.Now().UnixNano())
	dstore, err := leveldb.NewDatastore("", nil)

	//srcStore := dssync.MutexWrap(datastore.NewMapDatastore())
	srcStore := dssync.MutexWrap(dstore)
	h, _ := libp2p.New(context.Background(),
		libp2p.Identity(PrivKey),
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
	mockTasks := task.GenMockTask(mockTasksNum)
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

	time.Sleep(time.Second * 7)
	lp.Close()
	return
}
