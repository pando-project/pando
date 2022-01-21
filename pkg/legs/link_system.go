package legs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-graphsync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
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
			origBuf := buf.Bytes()

			log := log.With("cid", c)

			// Decode the node to check its type.
			n, err := decodeIPLDNode(buf)
			if err != nil {
				log.Errorw("Error decoding IPLD node in linksystem", "err", err)
				return errors.New("bad ipld data")
			}
			// If it is an advertisement.
			if isAdvertisement(n) {
				log.Infow("Received advertisement")

				// Verify that the signature is correct and the advertisement
				// is valid.
				ad, provID, err := verifyAdvertisement(n)
				if err != nil {
					return err
				}

				// Register provider or update existing registration.  The
				// provider must be allowed by policy to be registered.
				err = reg.RegisterOrUpdate(lctx.Ctx, provID, addrs, c)
				if err != nil {
					return err
				}

				// Store entries link into the reverse map so there is a way of
				// identifying what advertisementID announced these entries
				// when we come across the link.
				log.Debug("Saving map of entries to advertisement and advertisement data")
				elnk, err := ad.FieldEntries().AsLink()
				if err != nil {
					log.Errorw("Error getting link for entries from advertisement", "err", err)
					return errBadAdvert
				}
				err = putCidToAdMapping(lctx.Ctx, ds, elnk, c)
				if err != nil {
					log.Errorw("Error storing reverse map for entries in datastore", "err", err)
					return errors.New("cannot process advertisement")
				}

				// Persist the advertisement.  This is read later when
				// processing each chunk of entries, to get info common to all
				// entries in a chunk.
				return ds.Put(lctx.Ctx, dsKey(c.String()), origBuf)
			}
			log.Debug("Received IPLD node")
			// Any other type of node (like entries) are stored right away.
			return ds.Put(lctx.Ctx, dsKey(c.String()), origBuf)
		}, nil
	}
	//lsys.StorageWriteOpener = func(lnkCtx ipld.LinkContext) (io.Writer, ipld.BlockWriteCommitter, error) {
	//	var buffer settableBuffer
	//	committer := func(lnk ipld.Link) error {
	//		asCidLink, ok := lnk.(cidlink.Link)
	//		if !ok {
	//			return fmt.Errorf("unsupported link types")
	//		}
	//		fmt.Printf("[link-sys committer]time: %v, cid: %s\n", time.Now(), asCidLink.Cid)
	//		block, err := blocks.NewBlockWithCid(buffer.Bytes(), asCidLink.Cid)
	//		if err != nil {
	//			return err
	//		}
	//		err = bs.Put(lnkCtx.Ctx, block)
	//		return err
	//	}
	//	return &buffer, committer, nil
	//}
	return lsys
}

func decodeAd(n ipld.Node) (schema2.Advertisement, error) {
	nb := schema2.Type.Advertisement.NewBuilder()
	err := nb.AssignNode(n)
	if err != nil {
		return nil, err
	}
	return nb.Build().(schema2.Advertisement), nil
}

func verifyAdvertisement(n ipld.Node) (schema2.Advertisement, peer.ID, error) {
	ad, err := decodeAd(n)
	if err != nil {
		log.Errorw("Cannot decode advertisement", "err", err)
		return nil, peer.ID(""), errBadAdvert
	}
	// Verify advertisement signature
	signerID, err := schema2.VerifyAdvertisement(ad)
	if err != nil {
		// stop exchange, verification of signature failed.
		log.Errorw("Advertisement signature verification failed", "err", err)
		return nil, peer.ID(""), errInvalidAdvertSignature
	}

	// Get provider ID from advertisement.
	provID, err := providerFromAd(ad)
	if err != nil {
		log.Errorw("Cannot get provider from advertisement", "err", err)
		return nil, peer.ID(""), errBadAdvert
	}

	// Verify that the advertised provider has signed, and
	// therefore approved, the advertisement regardless of who
	// published the advertisement.
	if signerID != provID {
		// TODO: Have policy that allows a signer (publisher) to
		// sign advertisements for certain providers.  This will
		// allow that signer to add, update, and delete indexed
		// content on behalf of those providers.
		log.Errorw("Advertisement not signed by provider", "provider", provID, "signer", signerID)
		return nil, peer.ID(""), errInvalidAdvertSignature
	}

	return ad, provID, nil
}

// providerFromAd reads the provider ID from an advertisement
func providerFromAd(ad schema2.Advertisement) (peer.ID, error) {
	provider, err := ad.FieldProvider().AsString()
	if err != nil {
		return peer.ID(""), fmt.Errorf("cannot read provider from advertisement: %s", err)
	}

	providerID, err := peer.Decode(provider)
	if err != nil {
		return peer.ID(""), fmt.Errorf("cannot decode provider peer id: %s", err)
	}

	return providerID, nil
}

// decodeIPLDNode decodes an ipld.Node from bytes read from an io.Reader.
func decodeIPLDNode(r io.Reader) (ipld.Node, error) {
	// NOTE: Considering using the schema prototypes.  This was failing, using
	// a map gives flexibility.  Maybe is worth revisiting this again in the
	// future.
	nb := basicnode.Prototype.Any.NewBuilder()
	err := dagjson.Decode(nb, r)
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
func (c *Core) storageHook() graphsync.OnIncomingBlockHook {
	return func(p peer.ID, responseData graphsync.ResponseData, blockData graphsync.BlockData, hookActions graphsync.IncomingBlockHookActions) {
		log.Debug("hook - Triggering after a block has been stored")
		// Get cid of the node received.
		cid := blockData.Link().(cidlink.Link).Cid

		// Get entries node from datastore.
		_, err := c.BS.Get(context.Background(), cid)
		if err != nil {
			log.Errorf("Error while fetching the node from datastore: %s", err)
			return
		}

		log.Debugf("[recv] block from graphysnc.cid %s\n", cid.String())
	}
}

// Checks if an IPLD node is an advertisement, by looking to see if it has a
// "Signature" field.  We may need additional checks if we extend the schema
// with new types that are traversable.
func isAdvertisement(n ipld.Node) bool {
	indexID, _ := n.LookupByString("Signature")
	return indexID != nil
}
