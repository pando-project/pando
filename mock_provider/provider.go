package main

import (
	"context"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"time"

	"bytes"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"io"

	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/libp2p/go-libp2p"
	"github.com/multiformats/go-multicodec"
)

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

func mkLinkSystem(ds datastore.Batching) ipld.LinkSystem {
	lsys := cidlink.DefaultLinkSystem()
	lsys.StorageReadOpener = func(_ ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		c := lnk.(cidlink.Link).Cid
		val, err := ds.Get(datastore.NewKey(c.String()))
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(val), nil
	}
	lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			return ds.Put(datastore.NewKey(c.String()), buf.Bytes())
		}, nil
	}
	return lsys
}

func main() {
	ds := dssync.MutexWrap(datastore.NewMapDatastore())
	h, err := libp2p.New(context.Background())
	if err != nil {
		panic(err)
	}
	linkSys := mkLinkSystem(ds)

	lp, err := golegs.NewPublisher(context.Background(), h, ds, linkSys, "pppp")

	if err != nil {
		panic(err)
	}

	itm1 := basicnode.NewString("dsadasdasssda")
	lnk1, err := mkRoot(ds, itm1)
	if err != nil {
		panic(err)
	}

	if err := lp.UpdateRoot(context.Background(), lnk1.(cidlink.Link).Cid); err != nil {
		panic(err)
	}

	lp.Close()
	time.Sleep(time.Second * 5)
}
