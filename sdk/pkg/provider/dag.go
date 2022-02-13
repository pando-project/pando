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
	"pando/pkg/types/schema"

	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	link "pando/pkg/legs"
	"pando/sdk/pkg"
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

func NewDAGProvider(privateKeyStr string, connectTimeout time.Duration, pushTimeout time.Duration) (*DAGProvider, error) {
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

	storageCore.LinkSys = link.MkLinkSystem(storageCore.Blockstore, nil)
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

func (p *DAGProvider) Push(metadata schema.Metadata) (cid.Cid, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.PushTimeout)
	defer cancel()

	// Store the metadata locally.
	c, err := p.PushLocal(ctx, metadata)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to publish meta data locally: %s", err)
	}

	logger.Infow("Publishing meta data in pubsub channel", "cid", c)
	// Publish the metadata.
	err = p.LegsPublisher.UpdateRoot(ctx, c)
	if err != nil {
		return cid.Undef, err
	}
	return c, nil
}

func (p *DAGProvider) PushLocal(ctx context.Context, metadata schema.Metadata) (cid.Cid, error) {
	metadataLink, err := schema.MetadataLink(p.Core.LinkSys, metadata)
	if err != nil {
		return cid.Undef, fmt.Errorf("cannot generate metadata link: %s", err)
	}

	c := metadataLink.ToCid()

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
