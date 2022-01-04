package sdk

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
)

type core struct {
	mutexDatastore *datastoreSync.MutexDatastore
	blockstore     blockstore.Blockstore
	dagService     ipldFormat.DAGService
}

type DAGProvider struct {
	Host          host.Host
	PrivateKey    crypto.PrivKey
	LegsPublisher legs.LegPublisher
	Core          *core
}

func NewDAGProvider(privateKeyStr string) (*DAGProvider, error) {
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

	return &DAGProvider{
		Host:          providerHost,
		PrivateKey:    privateKey,
		LegsPublisher: legsPublisher,
		Core:          storageCore,
	}, nil
}

func (d *DAGProvider) ConnectPando(peerAddress string, peerID string) error {
	pandoPeerInfo, err := NewPandoPeerInfo(peerAddress, peerID)
	if err != nil {
		return err
	}

	return d.Host.Connect(context.Background(), *pandoPeerInfo)
}

func (d *DAGProvider) Close() error {
	return d.LegsPublisher.Close()
}

func (d *DAGProvider) Push(node ipldFormat.Node) error {
	return d.LegsPublisher.UpdateRoot(context.Background(), node.Cid())
}
