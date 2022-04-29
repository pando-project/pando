package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/filecoin-project/go-legs"
	"github.com/filecoin-project/go-legs/dtsync"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/kenlabs/PandoStore/pkg/config"
	"github.com/kenlabs/PandoStore/pkg/store"
	"github.com/kenlabs/pando/pkg/types/schema"

	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	link "github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/sdk/pkg"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"time"
)

type core struct {
	MutexDatastore *datastoreSync.MutexDatastore
	Blockstore     blockstore.Blockstore
	LinkSys        ipld.LinkSystem
}

type DAGProvider struct {
	Host           host.Host
	PrivateKey     crypto.PrivKey
	LegsPublisher  legs.Publisher
	Core           *core
	ConnectTimeout time.Duration
	PushTimeout    time.Duration
}

const topic = "/pando/v0.0.1"

const latestMedataKey = "/sync/metadata"

var dsLatestMetadataKey = datastore.NewKey(latestMedataKey)

var logger = log.Logger("sdk-provider-DAG")

func NewMetaProvider(privateKeyStr string, connectTimeout time.Duration, pushTimeout time.Duration) (*DAGProvider, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	providerHost, err := libp2p.New(libp2p.Identity(privateKey))
	if err != nil {
		return nil, err
	}

	storageCore := &core{}
	datastore, err := leveldb.NewDatastore("", nil)
	storageCore.MutexDatastore = datastoreSync.MutexWrap(datastore)
	storageCore.Blockstore = blockstore.NewBlockstore(storageCore.MutexDatastore)
	ps, err := store.NewStoreFromDatastore(context.Background(), storageCore.MutexDatastore, &config.StoreConfig{
		SnapShotInterval: "9999h",
	})
	if err != nil {
		return nil, err
	}

	storageCore.LinkSys = link.MkLinkSystem(ps, nil, nil)
	legsPublisher, err := dtsync.NewPublisher(providerHost, datastore, storageCore.LinkSys, topic)

	time.Sleep(2 * time.Second)

	return &DAGProvider{
		Host:           providerHost,
		PrivateKey:     privateKey,
		LegsPublisher:  legsPublisher,
		Core:           storageCore,
		ConnectTimeout: connectTimeout,
		PushTimeout:    pushTimeout,
	}, nil
}

func (p *DAGProvider) ConnectPando(peerAddress string, peerID string) error {
	pandoPeerInfo, err := pkg.NewPandoPeerInfo(peerAddress, peerID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.ConnectTimeout)
	defer cancel()

	return p.Host.Connect(ctx, *pandoPeerInfo)
}

func (p *DAGProvider) Close() error {
	return p.LegsPublisher.Close()
}

func (p *DAGProvider) NewMetadata(payload []byte) (*schema.Metadata, error) {
	return schema.NewMetaWithBytesPayload(payload, p.Host.ID(), p.PrivateKey)
}

func (p *DAGProvider) NewMetadataWithLink(payload []byte, link datamodel.Link) (*schema.Metadata, error) {
	return schema.NewMetadataWithLink(payload, p.Host.ID(), p.PrivateKey, link)
}

//func (p *DAGProvider) AppendMetadata(metadata *schema.Metadata, payload []byte) (*schema.Metadata, error) {
//	previousID, err := p.PushLocal(context.Background(), metadata)
//	if err != nil {
//		return nil, err
//	}
//	return metadata.AppendMetadata(previousID, p.Host.ID(), payload, p.PrivateKey)
//}

func (p *DAGProvider) Push(metadata *schema.Metadata) (cid.Cid, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.PushTimeout)
	defer cancel()

	// Store the metadata locally.
	c, err := p.PushLocal(ctx, metadata)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to publish metadata locally: %s", err)
	}

	logger.Infow("Publishing metadata in pubsub channel", "cid", c)
	// Publish the metadata.
	err = p.LegsPublisher.UpdateRoot(ctx, c)
	if err != nil {
		return cid.Undef, err
	}
	return c, nil
}

func (p *DAGProvider) PushLocal(ctx context.Context, metadata *schema.Metadata) (cid.Cid, error) {
	metadataLink, err := schema.MetadataLink(p.Core.LinkSys, metadata)
	if err != nil {
		return cid.Undef, fmt.Errorf("cannot generate metadata link: %s", err)
	}

	c := metadataLink.(cidlink.Link).Cid

	logger.Infow("Storing metadata locally", "cid", c.String())
	err = p.putLatestMetadata(ctx, c.Bytes())
	if err != nil {
		return cid.Undef, fmt.Errorf("cannot store latest metadata in blockstore: %s", err)
	}
	return c, nil
}

func (p *DAGProvider) putLatestMetadata(ctx context.Context, metadataID []byte) error {
	return p.Core.MutexDatastore.Put(ctx, dsLatestMetadataKey, metadataID)
}
