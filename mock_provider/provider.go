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

var mockTasksNum = 5

var PandoAddrStr = "/ip4/192.168.1.172/tcp/5003"

var TestProviderIdentity = &config.Identity{
	PeerID:  "12D3KooWDi135q9xcE7xiRN1bBZZGc15dSyFRgm7pajTmt7ndCX5",
	PrivKey: "CAESQHMFRinebmZ/C2zo8tJfYlWxrW5jUIaNoKndLO/LNuLlOc1eZZUi3InQk7QIx0ggEBtkisx7wd+bFsYJrjkc2Uw=",
}

var PrivKey, _ = TestProviderIdentity.DecodePrivateKey("")

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

//func getDagNodes()[]ipld.Node{
//	a := merkledag.NewRawNode([]byte("aaaa"))
//	b := merkledag.NewRawNode([]byte("bbbb"))
//	c := merkledag.NewRawNode([]byte("cccc"))
//
//	nd1 := &merkledag.ProtoNode{}
//	nd1.AddNodeLink("cat", a)
//
//	nd2 := &merkledag.ProtoNode{}
//	nd2.AddNodeLink("first", nd1)
//	nd2.AddNodeLink("dog", b)
//
//	nd3 := &merkledag.ProtoNode{}
//	nd3.AddNodeLink("second", nd2)
//	nd3.AddNodeLink("bear", c)
//}

func main() {
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

	ma, err := multiaddr.NewMultiaddr(PandoAddrStr + "/ipfs/12D3KooWHAVmpVPYMMS1tkYJZ5v5WHYrKNDm6BmR643EtYXqnyRK")
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
	time.Sleep(time.Second * 3)
	lnks := make([]ipld.Link, 0)
	for i := 0; i < len(mockTasks); i++ {
		lk, err := Store(bs, nodes[i])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Add cid %s to root\r\n", lk.String())
		lnks = append(lnks, lk)
	}

	for i := 0; i < len(mockTasks); i++ {
		err = lp.UpdateRoot(context.Background(), lnks[i].(cidlink.Link).Cid)
		if err != nil {
			log.Fatal(err)
		}
	}

	time.Sleep(time.Second * 5)
	lp.Close()
	return
}

//func disableEscapeHtml(data interface{}) (string, error) {
//	bf := bytes.NewBuffer([]byte{})
//	jsonEncoder := json.NewEncoder(bf)
//	jsonEncoder.SetEscapeHTML(false)
//	if err := jsonEncoder.Encode(data); err != nil {
//		return "", err
//	}
//	return bf.String(), nil
//}
////func main(){
////	s :=
//
//
//
//}
