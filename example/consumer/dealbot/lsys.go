package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"io"
	"time"
)

// decodeIPLDNode decodes an ipld.Node from bytes read from an io.Reader.
func decodeIPLDNode(codec uint64, r io.Reader, prototype ipld.NodePrototype) (ipld.Node, error) {
	// NOTE: Considering using the schema prototypes.  This was failing, using
	// a map gives flexibility.  Maybe is worth revisiting this again in the
	// future.
	nb := prototype.NewBuilder()
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

func isMetadata(n ipld.Node) bool {
	signature, _ := n.LookupByString("Signature")
	provider, _ := n.LookupByString("Provider")
	payload, _ := n.LookupByString("Payload")
	return signature != nil && provider != nil && payload != nil
}

func MkLinkSystem(bs blockstore.Blockstore, ch chan Status) ipld.LinkSystem {
	log := logging.Logger("consumer-lsys")
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
			n, err := decodeIPLDNode(codec, buf, basicnode.Prototype.Any)
			if err != nil {
				log.Errorw("Error decoding IPLD node in linksystem", "err", err)
				return errors.New("bad ipld data")
			}
			//fmt.Println("Reveiving ipld node:")
			t, err := UnwrapFinishedTask(n)
			if err == nil {
				//dagjson.Encode(n, os.Stdout)
				//fmt.Println(t.Status)
				if ch != nil {
					go func() {
						ctx, cncl := context.WithTimeout(context.Background(), time.Second*3)
						defer cncl()
						select {
						case _ = <-ctx.Done():
							log.Errorf("time out for send status info")
							return
						case ch <- t.Status:
						}
					}()
				}
			} else {
				log.Debugf("not FinishedTask, ignore...")
			}

			//dagjson.Encode(n, os.Stdout)
			if isMetadata(n) {
				log.Infow("Received metadata")
				// todo:  how to deal different signature version
				//_, peerid, err := verifyMetadata(n)
				//if err != nil {
				//	return err
				//}
				block, err := blocks.NewBlockWithCid(origBuf, c)
				if err != nil {
					return err
				}
				return bs.Put(lctx.Ctx, block)
			}
			block, err := blocks.NewBlockWithCid(origBuf, c)
			if err != nil {
				return err
			}
			log.Debugf("Received unexpected IPLD node, cid: %s", c.String())
			return bs.Put(lctx.Ctx, block)
		}, nil
	}
	return lsys
}
