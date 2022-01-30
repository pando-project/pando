package legs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	schema2 "pando/types/schema"

	// dagjson codec registered for encoding

	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func MkLinkSystem(bs blockstore.Blockstore) ipld.LinkSystem {
	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	lsys.StorageReadOpener = func(lnkCtx ipld.LinkContext, lnk ipld.Link) (io.Reader, error) {
		asCidLink, ok := lnk.(cidlink.Link)
		if !ok {
			return nil, fmt.Errorf("unsupported link types")
		}
		block, err := bs.Get(lnkCtx.Ctx, asCidLink.Cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(block.RawData()), nil
	}
	lsys.StorageWriteOpener = func(lctx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
		buf := bytes.NewBuffer(nil)
		return buf, func(lnk ipld.Link) error {
			c := lnk.(cidlink.Link).Cid
			codec := lnk.(cidlink.Link).Prefix().Codec
			origBuf := buf.Bytes()

			log := log.With("cid", c)

			// Decode the node to check its type.
			n, err := decodeIPLDNode(codec, buf)
			if err != nil {
				log.Errorw("Error decoding IPLD node in linksystem", "err", err)
				return errors.New("bad ipld data")
			}
			// If it is an advertisement.
			if isMetadata(n) {
				log.Infow("Received advertisement")
				block, err := blocks.NewBlockWithCid(origBuf, c)
				if err != nil {
					return err
				}
				return bs.Put(lctx.Ctx, block)
			}
			log.Debug("Received unexpected IPLD node, skip")
			// Any other type of node (like entries) are stored right away.
			return nil
		}, nil
	}
	return lsys
}

func decodeAd(n ipld.Node) (schema2.Metadata, error) {
	nb := schema2.Type.Metadata.NewBuilder()
	err := nb.AssignNode(n)
	if err != nil {
		return nil, err
	}
	return nb.Build().(schema2.Metadata), nil
}

// decodeIPLDNode decodes an ipld.Node from bytes read from an io.Reader.
func decodeIPLDNode(codec uint64, r io.Reader) (ipld.Node, error) {
	// NOTE: Considering using the schema prototypes.  This was failing, using
	// a map gives flexibility.  Maybe is worth revisiting this again in the
	// future.
	nb := basicnode.Prototype.Any.NewBuilder()
	decoder, err := multicodec.LookupDecoder(codec)
	if err != nil {
		return nil, err
	}
	err = decoder(nb, r)
	if err != nil {
		return nil, err
	}
	return nb.Build(), nil
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
// graph sync.
//
// When we receive a block, if it is not an advertisement it means that we
// finished storing the list of entries of the advertisement, so we are ready
// to process them and ingest into the indexer core.
//func (c *Core) storageHook() graphsync.OnIncomingBlockHook {
//	return func(p peer.ID, responseData graphsync.ResponseData, blockData graphsync.BlockData, hookActions graphsync.IncomingBlockHookActions) {
//		log.Debug("hook - Triggering after a block has been stored")
//		// Get cid of the node received.
//		cid := blockData.Link().(cidlink.Link).Cid
//
//		// Get entries node from datastore.
//		_, err := c.BS.Get(context.Background(), cid)
//		if err != nil {
//			log.Errorf("Error while fetching the node from blockstore: %s", err)
//			return
//		}
//
//		// Decode block to IPLD node
//		node, err := decodeIPLDNode(c.Prefix().Codec, bytes.NewBuffer(val))
//		if err != nil {
//			log.Errorw("Error decoding ipldNode", "err", err)
//			return
//		}
//
//		log.Debugf("[recv] block from graphysnc.cid %s\n", cid.String())
//	}
//}

func (c *Core) storageHook(pubID peer.ID, cc cid.Cid) {
	log := log.With("publisher", pubID, "cid", cc)

	// Get data corresponding to the block.
	val, err := c.BS.Get(context.Background(), cc)
	if err != nil {
		log.Errorw("Error while fetching the node from datastore", "err", err)
		return
	}

	// Decode block to IPLD node
	node, err := decodeIPLDNode(cc.Prefix().Codec, bytes.NewBuffer(val.RawData()))
	if err != nil {
		log.Errorw("Error decoding ipldNode", "err", err)
		return
	}

	// If this is an advertisement, sync entries within it.
	if isMetadata(node) {
		ad, err := decodeAd(node)
		if err != nil {
			log.Errorw("Error decoding advertisement", "err", err)
			return
		}

		var prevCid cid.Cid
		if ad.FieldPreviousID().Exists() {
			lnk, err := ad.FieldPreviousID().Must().AsLink()
			if err != nil {
				log.Errorw("Cannot read previous link from metadata", "err", err)
			} else {
				prevCid = lnk.(cidlink.Link).Cid
			}
		}

		log.Infow("Incoming block is a metadata", "prevAd", prevCid)

		//go c.syncChainMeta(pubID, ad, cc, prevCid)
		return
	}
}

//func  (c *Core) syncChainMeta(from peer.ID, ad schema2.Metadata, adCid, prevCid cid.Cid){
//}

// Checks if an IPLD node is a Metadata, by looking to see if it has a
// "PreviousID" field.  We may need additional checks if we extend the schema
// with new types that are traversable.
func isMetadata(n ipld.Node) bool {
	prev, _ := n.LookupByString("PreviousID")
	return prev != nil
}
