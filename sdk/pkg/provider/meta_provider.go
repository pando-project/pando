package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	dt "github.com/filecoin-project/go-data-transfer"
	dtImpl "github.com/filecoin-project/go-data-transfer/impl"
	dtnetwork "github.com/filecoin-project/go-data-transfer/network"
	gstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	"github.com/filecoin-project/go-legs"
	"github.com/filecoin-project/go-legs/dtsync"
	"github.com/go-resty/resty/v2"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	gsimpl "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"
	"github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/kenlabs/pando-store/pkg/config"
	"github.com/kenlabs/pando-store/pkg/store"
	store2 "github.com/kenlabs/pando-store/pkg/types/store"
	"github.com/libp2p/go-libp2p"
	"github.com/pando-project/pando/pkg/types/schema"
	"net/http"
	"net/url"

	datastoreSync "github.com/ipfs/go-datastore/sync"
	dtSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	link "github.com/pando-project/pando/pkg/legs"
	"github.com/pando-project/pando/sdk/pkg"
	"time"
)

type core struct {
	MutexDatastore *datastoreSync.MutexDatastore
	BlockStore     blockstore.Blockstore
	LinkSys        ipld.LinkSystem
	PandoStore     *store.PandoStore
}

type MetaProvider struct {
	Host           host.Host
	PrivateKey     crypto.PrivKey
	LegsPublisher  legs.Publisher
	DtManager      dt.Manager
	Core           *core
	HttpClient     *resty.Client
	ConnectTimeout time.Duration
	PushTimeout    time.Duration
}

const topic = "/pando/v0.0.1"

const latestMedataKey = "/sync/metadata"

var dsLatestMetadataKey = datastore.NewKey(latestMedataKey)

var logger = log.Logger("sdk-provider-DAG")

func NewMetaProvider(privateKeyStr string, pandoAPI string, connectTimeout time.Duration, pushTimeout time.Duration) (*MetaProvider, error) {
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
	ds, err := leveldb.NewDatastore("", nil)
	storageCore.MutexDatastore = datastoreSync.MutexWrap(ds)
	storageCore.BlockStore = blockstore.NewBlockstore(storageCore.MutexDatastore)
	ps, err := store.NewStoreFromDatastore(context.Background(), storageCore.MutexDatastore, &config.StoreConfig{
		SnapShotInterval: "9999h",
		CacheSize:        config.DefaultCacheSize,
	})
	if err != nil {
		return nil, err
	}
	storageCore.PandoStore = ps

	storageCore.LinkSys = link.MkLinkSystem(ps, nil, nil)
	//legsPublisher, err := dtsync.NewPublisher(providerHost, storageCore.MutexDatastore, storageCore.LinkSys, topic)
	_, err = url.Parse(pandoAPI)
	if err != nil {
		return nil, err
	}
	httpClient := resty.New().SetBaseURL(pandoAPI).SetTimeout(connectTimeout).SetDebug(false)
	time.Sleep(2 * time.Second)

	dtNet := dtnetwork.NewFromLibp2pHost(providerHost)
	gsNet := gsnet.NewFromLibp2pHost(providerHost)
	gs := gsimpl.New(context.Background(), gsNet, storageCore.LinkSys)
	tp := gstransport.NewTransport(providerHost.ID(), gs)
	dtManager, err := dtImpl.NewDataTransfer(dtSync.MutexWrap(datastore.NewMapDatastore()), dtNet, tp)
	legsPublisher, err := dtsync.NewPublisherFromExisting(dtManager, providerHost, topic, storageCore.LinkSys,
		dtsync.AllowPeer(func(id peer.ID) bool {
			return true
		}))
	err = dtManager.Start(context.Background())
	if err != nil {
		logger.Errorf("start datasync manager failed. %s", err.Error())
	}

	return &MetaProvider{
		Host:           providerHost,
		PrivateKey:     privateKey,
		LegsPublisher:  legsPublisher,
		DtManager:      dtManager,
		HttpClient:     httpClient,
		Core:           storageCore,
		ConnectTimeout: connectTimeout,
		PushTimeout:    pushTimeout,
	}, nil
}

func (p *MetaProvider) ConnectPando(peerAddress string, peerID string) error {
	pandoPeerInfo, err := pkg.NewPandoPeerInfo(peerAddress, peerID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), p.ConnectTimeout)
	defer cancel()

	return p.Host.Connect(ctx, *pandoPeerInfo)
}

func (p *MetaProvider) Close() error {
	return p.LegsPublisher.Close()
}

func (p *MetaProvider) NewMetadata(payload []byte) (*schema.Metadata, error) {
	return schema.NewMetaWithBytesPayload(payload, p.Host.ID(), p.PrivateKey)
}

func (p *MetaProvider) NewMetadataWithPayload(payload datamodel.Node, cache *bool, collection *string) (*schema.Metadata, error) {
	return schema.NewMetaWithPayloadNode(payload, p.Host.ID(), p.PrivateKey, nil, cache, collection)
}

func (p *MetaProvider) NewMetadataWithPayloadLink(payload datamodel.Node, cache *bool, collection *string, link datamodel.Link) (*schema.Metadata, error) {
	return schema.NewMetaWithPayloadNode(payload, p.Host.ID(), p.PrivateKey, &link, cache, collection)
}

func (p *MetaProvider) NewMetadataWithLink(payload []byte, link datamodel.Link) (*schema.Metadata, error) {
	return schema.NewMetadataWithLink(payload, p.Host.ID(), p.PrivateKey, link)
}

func (p *MetaProvider) Push(metadata schema.Meta) (cid.Cid, error) {
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

func (p *MetaProvider) PushLocal(ctx context.Context, metadata schema.Meta) (cid.Cid, error) {
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

func (p *MetaProvider) putLatestMetadata(ctx context.Context, metadataID []byte) error {
	return p.Core.MutexDatastore.Put(ctx, dsLatestMetadataKey, metadataID)
}

type responseJson struct {
	Code    int                                      `json:"code"`
	Message string                                   `json:"message"`
	Data    struct{ Inclusion store2.MetaInclusion } `json:"Data"`
}

func (p *MetaProvider) CheckMetaState(ctx context.Context, c cid.Cid) (*store2.MetaInclusion, error) {
	res, err := pkg.HandleResError(p.HttpClient.R().Get("/metadata/inclusion?cid=" + c.String()))
	if err != nil {
		return nil, err
	}
	resJson := responseJson{}
	err = json.Unmarshal(res.Body(), &resJson)
	if err != nil {
		return nil, err
	}
	if resJson.Code != http.StatusOK {
		return nil, fmt.Errorf("error msg: %s", resJson.Message)
	}

	return &resJson.Data.Inclusion, nil
}
