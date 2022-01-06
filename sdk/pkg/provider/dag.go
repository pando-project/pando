package provider

import (
	"context"
	"encoding/base64"
	"github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-blockservice"
	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipldFormat "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	link "pando/pkg/legs"
	"pando/sdk/pkg"
	"time"
)

type core struct {
	mutexDatastore *datastoreSync.MutexDatastore
	blockstore     blockstore.Blockstore
	dagService     ipldFormat.DAGService
}

type DAGProvider struct {
	Host           host.Host
	PrivateKey     crypto.PrivKey
	LegsPublisher  legs.LegPublisher
	Core           *core
	ConnectTimeout time.Duration
	PushTimeout    time.Duration
}

func NewDAGProvider(privateKeyStr string, connectTimeout time.Duration, pushTimeout time.Duration) (*DAGProvider, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}
	privateKey, err := crypto.UnmarshalPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	providerHost, err := libp2p.New(context.Background(), libp2p.Identity(privateKey))
	if err != nil {
		return nil, err
	}

	storageCore := &core{}
	datastore, err := leveldb.NewDatastore("", nil)
	storageCore.mutexDatastore = datastoreSync.MutexWrap(datastore)
	storageCore.blockstore = blockstore.NewBlockstore(storageCore.mutexDatastore)
	storageCore.dagService = merkledag.NewDAGService(blockservice.New(
		storageCore.blockstore, offline.Exchange(storageCore.blockstore)))

	linkSys := link.MkLinkSystem(storageCore.blockstore)
	legsPublisher, err := legs.NewPublisher(context.Background(),
		providerHost, datastore, linkSys, "PandoPubSub")

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

func (p *DAGProvider) Push(node ipldFormat.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.PushTimeout)
	defer cancel()

	err := p.Core.dagService.Add(ctx, node)
	if err != nil {
		return err
	}
	return p.LegsPublisher.UpdateRoot(ctx, node.Cid())
}

func (p *DAGProvider) PushMany(nodes []ipldFormat.Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.PushTimeout)
	defer cancel()

	err := p.Core.dagService.AddMany(ctx, nodes)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		err := p.LegsPublisher.UpdateRoot(ctx, node.Cid())
		if err != nil {
			return err
		}
	}

	return nil
}
