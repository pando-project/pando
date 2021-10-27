package main

import (
	"Pando/legs"
	"Pando/mock_provider/task"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"io"
	"log"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
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

func mkRoot(srcStore datastore.Batching, n ipld.Node) (ipld.Link, error) {
	linkproto := cidlink.LinkPrototype{
		Prefix: cid.Prefix{
			Version:  1,
			Codec:    uint64(multicodec.DagJson),
			MhType:   uint64(multicodec.Sha2_256),
			MhLength: 16,
		},
	}
	lsys := cidlink.DefaultLinkSystem()
	lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return srcStore.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}

	return lsys.Store(ipld.LinkContext{}, linkproto, n)
}

func main() {
	srcStore := dssync.MutexWrap(datastore.NewMapDatastore())
	h, _ := libp2p.New(context.Background())
	fmt.Println("p2pHost addr:", h.Addrs())
	fmt.Println("p2pHost id:", h.ID())
	srcLnkS := legs.MkLinkSystem(srcStore)

	ma, err := multiaddr.NewMultiaddr("/ip4/192.168.1.172/tcp/5003/ipfs/QmZP28EezUd6Lsgxr7W9NxZ9JWFwzDmrntH6TsqFVnnz9z")
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

		nn := basicnode.NewBytes(taskBytes)
		nodes = append(nodes, nn)
		fmt.Println(taskBytes)
	}

	lp, err := golegs.NewPublisher(context.Background(), h, srcStore, srcLnkS, "pandotest")
	if err != nil {
		log.Fatal(err)
	}
	defer lp.Close()
	time.Sleep(time.Second * 3)
	lnks := make([]ipld.Link, 0)
	for i := 0; i < len(mockTasks); i++ {
		lk, err := mkRoot(srcStore, nodes[i])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Add cid %s to root", lk.String())
		lnks = append(lnks, lk)
	}
	time.Sleep(time.Second)

	for i := 0; i < len(mockTasks); i++ {
		if err := lp.UpdateRoot(context.Background(), lnks[i].(cidlink.Link).Cid); err != nil {
			log.Fatal(err)
		}
	}
	time.Sleep(time.Second * 10)

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
