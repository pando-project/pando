package legs

import (
	"bytes"
	"errors"
	"fmt"
	blocks "github.com/ipfs/go-block-format"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"pando/pkg/types/schema"

	// dagjson codec registered for encoding

	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func MkLinkSystem(bs blockstore.Blockstore, core *Core) ipld.LinkSystem {
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
				log.Infow("Received metadata")
				_, peerid, err := verifyMetadata(n)
				if err != nil {

					return err
				}
				block, err := blocks.NewBlockWithCid(origBuf, c)
				if err != nil {
					return err
				}
				if core != nil {
					go core.SendRecvMeta(c, peerid)
				}
				return bs.Put(lctx.Ctx, block)
			}
			log.Debug("Received unexpected IPLD node, skip")
			return nil
		}, nil
	}
	return lsys
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

// Checks if an IPLD node is a Metadata, by looking to see if it has a
// "PreviousID" field.  We may need additional checks if we extend the schema
// with new types that are traversable.
func isMetadata(n ipld.Node) bool {
	prev, _ := n.LookupByString("Payload")
	return prev != nil
}

func decodeAd(n ipld.Node) (schema.Metadata, error) {
	nb := schema.Type.Metadata.NewBuilder()
	err := nb.AssignNode(n)
	if err != nil {
		return nil, err
	}
	return nb.Build().(schema.Metadata), nil
}

func verifyMetadata(n ipld.Node) (schema.Metadata, peer.ID, error) {
	meta, err := decodeAd(n)
	if err != nil {
		log.Errorw("Cannot decode advertisement", "err", err)
		return nil, peer.ID(""), err
	}
	// Verify advertisement signature
	signerID, err := schema.VerifyMetadata(meta)
	if err != nil {
		// stop exchange, verification of signature failed.
		log.Errorw("Metadata signature verification failed", "err", err)
		return nil, peer.ID(""), err
	}

	// Get provider ID from advertisement.
	provID, err := providerFromAd(meta)
	if err != nil {
		log.Errorw("Cannot get provider from advertisement", "err", err)
		return nil, peer.ID(""), err
	}

	// Verify that the meta provider has signed, and
	// therefore approved, the metadata regardless of who
	// published the metadata.
	if signerID != provID {
		log.Errorw("Metadata not signed by provider", "provider", provID, "signer", signerID)
		return nil, peer.ID(""), err
	}

	return meta, provID, nil
}

// providerFromAd reads the provider ID from an advertisement
func providerFromAd(ad schema.Metadata) (peer.ID, error) {
	provider, err := ad.FieldProvider().AsString()
	if err != nil {
		return peer.ID(""), fmt.Errorf("cannot read provider from advertisement: %w", err)
	}

	providerID, err := peer.Decode(provider)
	if err != nil {
		return peer.ID(""), fmt.Errorf("cannot decode provider peer id: %w", err)
	}

	return providerID, nil
}
