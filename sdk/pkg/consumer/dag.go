package consumer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	dt "github.com/filecoin-project/go-data-transfer/impl"
	dtnetwork "github.com/filecoin-project/go-data-transfer/network"
	gstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	"github.com/filecoin-project/go-legs"
	"github.com/go-resty/resty/v2"
	"github.com/ipfs/go-cid"
	datastoreSync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	gsimpl "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipfs/go-log/v2"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	link "github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/sdk/pkg"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"net/http"
	"net/url"
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
	PandoPeerInfo  *peer.AddrInfo
	HttpClient     *resty.Client
	ConnectTimeout time.Duration
	SyncTimeout    time.Duration
}

const SubscribeTopic = "/pando/v0.0.1"

var logger = log.Logger("sdk-consumer-DAG")

func NewDAGConsumer(privateKeyStr string, pandoAPI string, connectTimeout time.Duration, syncTimeout time.Duration) (*DAGConsumer, error) {
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
	storageCore.LinkSys = link.MkLinkSystem(storageCore.Blockstore, nil, nil)

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
		nil,
		storageCore.LinkSys,
		SubscribeTopic,
		nil,
		legs.DtManager(dataManager, graphExchange),
	)
	if err != nil {
		return nil, err
	}

	_, err = url.Parse(pandoAPI)
	if err != nil {
		return nil, err
	}
	httpClient := resty.New().SetBaseURL(pandoAPI).SetTimeout(connectTimeout).SetDebug(false)

	return &DAGConsumer{
		Host:           consumerHost,
		PrivateKey:     privateKey,
		Core:           storageCore,
		Subscriber:     subscriber,
		HttpClient:     httpClient,
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

	return c.Host.Connect(ctx, *pandoPeerInfo)
}

func (c *DAGConsumer) Close() error {
	return c.Subscriber.Close()
}

type ResponseJson struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    struct{ Cid string } `json:"Data"`
}

func (c *DAGConsumer) GetLatestHead(providerPeerID string) (cid.Cid, error) {
	res, err := handleResError(c.HttpClient.R().Get("/provider/head?peerid=" + providerPeerID))
	if err != nil {
		return cid.Undef, err
	}
	resJson := ResponseJson{}
	err = json.Unmarshal(res.Body(), &resJson)
	if err != nil {
		return cid.Undef, err
	}
	if resJson.Code != http.StatusOK {
		return cid.Undef, fmt.Errorf("error msg: %s", resJson.Message)
	}

	nextCid, err := cid.Decode(resJson.Data.Cid)
	if err != nil {
		return cid.Undef, err
	}

	return nextCid, nil
}

func handleResError(res *resty.Response, err error) (*resty.Response, error) {
	errTmpl := "failed to get latest head, error: %v"
	if err != nil {
		return res, err
	}
	if res.IsError() {
		return res, fmt.Errorf(errTmpl, res.Error())
	}
	if res.StatusCode() != http.StatusOK {
		return res, fmt.Errorf(errTmpl, fmt.Sprintf("expect 200, got %d", res.StatusCode()))
	}

	return res, nil
}

func (c *DAGConsumer) GetLatestSync() cid.Cid {
	lnk := c.Subscriber.GetLatestSync(c.PandoPeerInfo.ID)
	cidLink, ok := lnk.(cidlink.Link)
	if !ok {
		logger.Errorf("lnk does not supported: %s", lnk.String())
	}
	return cidLink.Cid
}

func (c *DAGConsumer) Sync(nextCid cid.Cid, selector ipld.Node) (cid.Cid, error) {
	ctx, syncCancel := context.WithTimeout(context.Background(), c.SyncTimeout)
	defer syncCancel()

	return c.Subscriber.Sync(ctx, c.PandoPeerInfo.ID, nextCid, selector, nil)
}

func (c *DAGConsumer) Start(pandoAddr string, pandoPeerID string, providerPeerID string) error {
	logging.SetAllLoggers(logging.LevelDebug)
	//err := logging.SetLogLevel("addrutil", "warn")
	//if err != nil {
	//	return err
	//}

	err := c.ConnectPando(pandoAddr, pandoPeerID)
	if err != nil {
		return err
	}
	headCid, err := c.GetLatestHead(providerPeerID)
	if err != nil {
		return err
	}
	latestSyncCid, err := c.Sync(headCid, nil)
	fmt.Println("cid: ", latestSyncCid)
	if err != nil {
		return err
	}
	fmt.Println("sync succeed")

	return nil
}
