package consumer

import (
	"context"
	"encoding/base64"
	dt "github.com/filecoin-project/go-data-transfer/impl"
	dtnetwork "github.com/filecoin-project/go-data-transfer/network"
	gstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	"github.com/filecoin-project/go-legs"
	legsSync "github.com/filecoin-project/go-legs/dtsync"
	"github.com/ipfs/go-cid"
	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	gsimpl "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	link "github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/sdk/pkg"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

type core struct {
	Blockstore     blockstore.Blockstore
	MutexDatastore *datastoreSync.MutexDatastore
	LinkSys        ipld.LinkSystem
}

type DAGConsumer struct {
	Host           host.Host
	PrivateKey     crypto.PrivKey
	Core           *core
	Subscriber     *legs.Subscriber
	LegsSync       *legsSync.Sync
	LegsSyncer     *legsSync.Syncer
	PandoPeerInfo  *peer.AddrInfo
	ConnectTimeout time.Duration
	SyncTimeout    time.Duration
}

const SubscribeTopic = "/pando/v0.0.1"

var logger = log.Logger("sdk-consumer-DAG")

func NewDAGConsumer(privateKeyStr string, connectTimeout time.Duration, syncTimeout time.Duration) (*DAGConsumer, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	consumerHost, err := libp2p.New(libp2p.Identity(privateKey))
	if err != nil {
		return nil, err
	}

	storageCore := &core{}
	datastore, err := leveldb.NewDatastore("", nil)
	storageCore.MutexDatastore = datastoreSync.MutexWrap(datastore)
	storageCore.Blockstore = blockstore.NewBlockstore(storageCore.MutexDatastore)
	storageCore.LinkSys = link.MkLinkSystem(storageCore.Blockstore, nil)

	graphSyncNet := gsnet.NewFromLibp2pHost(consumerHost)
	dataTransferNet := dtnetwork.NewFromLibp2pHost(consumerHost)
	graphExchange := gsimpl.New(context.Background(), graphSyncNet, storageCore.LinkSys)
	graphTransport := gstransport.NewTransport(consumerHost.ID(), graphExchange)

	dataManager, err := dt.NewDataTransfer(storageCore.MutexDatastore, dataTransferNet, graphTransport)
	if err != nil {
		return nil, err
	}
	err = dataManager.Start(context.Background())
	if err != nil {
		return nil, err
	}

	subscriber, err := legs.NewSubscriber(consumerHost,
		storageCore.MutexDatastore,
		storageCore.LinkSys,
		SubscribeTopic,
		nil,
		legs.DtManager(dataManager),
	)

	sync, err := legsSync.NewSyncWithDT(consumerHost, dataManager)

	return &DAGConsumer{
		Host:           consumerHost,
		PrivateKey:     privateKey,
		Core:           storageCore,
		Subscriber:     subscriber,
		LegsSync:       sync,
		ConnectTimeout: connectTimeout,
		SyncTimeout:    syncTimeout,
	}, nil
}

func (c *DAGConsumer) ConnectPando(peerAddress string, peerID string) error {
	pandoPeerInfo, err := pkg.NewPandoPeerInfo(peerAddress, peerID)
	if err != nil {
		return err
	}
	c.PandoPeerInfo = pandoPeerInfo

	ctx, cancel := context.WithTimeout(context.Background(), c.ConnectTimeout)
	defer cancel()

	c.LegsSyncer = c.LegsSync.NewSyncer(pandoPeerInfo.ID, SubscribeTopic)

	return c.Host.Connect(ctx, *pandoPeerInfo)
}

func (c *DAGConsumer) Close() error {
	return c.Subscriber.Close()
}

func (c *DAGConsumer) GetLatestSync() cid.Cid {
	lnk := c.Subscriber.GetLatestSync(c.PandoPeerInfo.ID)
	cidLink, ok := lnk.(cidlink.Link)
	if !ok {
		logger.Errorf("lnk does not supported: %s", lnk.String())
	}
	return cidLink.Cid
}

func (c *DAGConsumer) Sync(nextCid cid.Cid, selector ipld.Node) error {
	ctx, syncCancel := context.WithTimeout(context.Background(), c.SyncTimeout)
	defer syncCancel()

	return c.LegsSyncer.Sync(ctx, nextCid, selector)
}
