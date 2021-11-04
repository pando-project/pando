package legs

import (
	"bytes"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"io"

	"github.com/ipld/go-ipld-prime"

	// dagjson codec registered for encoding

	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

//func dsKey(k string) datastore.Key {
//	return datastore.NewKey(k)
//}

func MkLinkSystem(bs blockstore.Blockstore) ipld.LinkSystem {
	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	lsys.StorageReadOpener = func(lnkCtx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		asCidLink, ok := lnk.(cidlink.Link)
		if !ok {
			return nil, fmt.Errorf("unsupported link type")
		}
		block, err := bs.Get(asCidLink.Cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(block.RawData()), nil
	}
	lsys.StorageWriteOpener = func(lnkCtx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		var buffer settableBuffer
		committer := func(lnk ipld.Link) error {
			asCidLink, ok := lnk.(cidlink.Link)
			if !ok {
				return fmt.Errorf("unsupported link type")
			}
			block, err := blocks.NewBlockWithCid(buffer.Bytes(), asCidLink.Cid)
			if err != nil {
				return err
			}
			return bs.Put(block)
		}
		return &buffer, committer, nil
	}
	return lsys
	//lsys.StorageReadOpener = func(_ ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
	//	c := lnk.(cidlink.Link).Cid
	//	val, err := ds.Get(datastore.NewKey(c.String()))
	//	if err != nil {
	//		return nil, err
	//	}
	//	return bytes.NewBuffer(val), nil
	//}
	//lsys.StorageWriteOpener = func(_ ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
	//	buf := bytes.NewBuffer(nil)
	//	return buf, func(lnk ipld.Link) error {
	//		c := lnk.(cidlink.Link).Cid
	//		return ds.Put(datastore.NewKey(c.String()), buf.Bytes())
	//	}, nil
	//}
}

type settableBuffer struct {
	bytes.Buffer
	didSetData bool
	data       []byte
}

func (sb *settableBuffer) SetBytes(data []byte) error {
	sb.didSetData = true
	sb.data = data
	return nil
}

func (sb *settableBuffer) Bytes() []byte {
	if sb.didSetData {
		return sb.data
	}
	return sb.Buffer.Bytes()
}

// storageHook determines the logic to run when a new block is received through
// graphsync.
//
// When we receive a block, if it is not an advertisement it means that we
// finished storing the list of entries of the advertisement, so we are ready
// to process them and ingest into the indexer core.
//func (i *LegsCore) storageHook() graphsync.OnIncomingBlockHook {
//	return func(p peer.ID, responseData graphsync.ResponseData, blockData graphsync.BlockData, hookActions graphsync.IncomingBlockHookActions) {
//		log.Debug("hook - Triggering after a block has been stored")
//		// Get cid of the node received.
//		c := blockData.Link().(cidlink.Link).Cid
//
//		// Get entries node from datastore.
//		val, err := i.DS.Get(dsKey(c.String()))
//		if err != nil {
//			log.Errorf("Error while fetching the node from datastore: %s", err)
//			return
//		}
//
//		// Decode entries into an IPLD node
//		nentries, err := decodeIPLDNode(bytes.NewBuffer(val))
//		if err != nil {
//			log.Errorf("Error decoding ipldNode: %s", err)
//			return
//		}
//
//		log.Debugf("[received] block from graphysnc.cid %s\r\n%s", c.String(), nentries.Kind())
//	}
//}

// decodeIPLDNode from a reader
// This is used to get the ipld.Node from a set of raw bytes.
func decodeIPLDNode(r io.Reader) (ipld.Node, error) {
	// NOTE: Considering using the schema prototypes.
	// This was failing, using a map gives flexibility.
	// Maybe is worth revisiting this again in the future.
	nb := basicnode.Prototype.Any.NewBuilder()
	err := dagjson.Decode(nb, r)
	if err != nil {
		return nil, err
	}
	return nb.Build(), nil
}
