package legs

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	dt "github.com/filecoin-project/go-data-transfer/impl"
	dtnetwork "github.com/filecoin-project/go-data-transfer/network"
	gstransport "github.com/filecoin-project/go-data-transfer/transport/graphsync"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	dtsync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-graphsync"
	gsimpl "github.com/ipfs/go-graphsync/impl"
	gsnet "github.com/ipfs/go-graphsync/network"
	"github.com/kenlabs/pando-store/pkg/store"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pando-project/pando/pkg/metadata"
	"github.com/pando-project/pando/pkg/metrics"
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/policy"
	"github.com/pando-project/pando/pkg/registry"
	"github.com/pando-project/pando/pkg/util/log"
	"strings"
	"sync"
	"time"
)

var logger = log.NewSubsystemLogger()

const (
	// SyncPrefix used to track the latest sync in datastore.
	SyncPrefix = "/sync/"
	// PubSubTopic used for legs transportation
	PubSubTopic = "/pando/v0.0.1"
)

type Core struct {
	Host              host.Host
	DS                datastore.Batching
	CS                *badger.DB
	PS                *store.PandoStore
	GS                graphsync.GraphExchange
	LS                *golegs.Subscriber
	reg               *registry.Registry
	cancelSyncFn      context.CancelFunc
	recvMetaCh        chan<- *metadata.MetaRecord
	backupGenInterval time.Duration
	rateLimiter       *policy.Limiter

	waitForPendingSyncs sync.WaitGroup
	watchDone           chan struct{}
	options             *option.DaemonOptions
}

func NewLegsCore(ctx context.Context,
	host host.Host,
	ds datastore.Batching,
	cs *badger.DB,
	ps *store.PandoStore,
	outMetaCh chan<- *metadata.MetaRecord,
	backupGenInterval time.Duration,
	rateLimiter *policy.Limiter, reg *registry.Registry, options *option.DaemonOptions) (*Core, error) {

	c := &Core{
		Host:              host,
		DS:                ds,
		CS:                cs,
		PS:                ps,
		reg:               reg,
		recvMetaCh:        outMetaCh,
		backupGenInterval: backupGenInterval,
		rateLimiter:       rateLimiter,
		watchDone:         make(chan struct{}),
		options:           options,
	}

	ls, gs, err := c.initSub(ctx, host, ds, ps, reg)
	if err != nil {
		return nil, fmt.Errorf("failed to create legs subscriber, err: %s", err.Error())
	}
	c.LS = ls
	c.GS = gs

	err = c.restoreLatestSync()
	if err != nil {
		_ = ls.Close()
		return nil, err
	}

	onSyncFin, cancelSyncFn := ls.OnSyncFinished()
	c.cancelSyncFn = cancelSyncFn

	go c.watchSyncFinished(onSyncFin)
	go c.autoSync()

	logger.Debugf("LegCore started and all hooks and linksystem registered")

	return c, nil
}

func (c *Core) initSub(ctx context.Context, h host.Host, ds datastore.Batching, ps *store.PandoStore, reg *registry.Registry) (*golegs.Subscriber, graphsync.GraphExchange, error) {
	lnkSys := MkLinkSystem(ps, c, reg)
	gsNet := gsnet.NewFromLibp2pHost(h)
	dtNet := dtnetwork.NewFromLibp2pHost(h)
	gs := gsimpl.New(context.Background(), gsNet, lnkSys)
	tp := gstransport.NewTransport(h.ID(), gs)

	dtManager, err := dt.NewDataTransfer(dtsync.MutexWrap(datastore.NewMapDatastore()), dtNet, tp)
	if err != nil {
		return nil, nil, err
	}
	err = dtManager.Start(ctx)
	if err != nil {
		return nil, nil, err
	}
	// todo
	//defer dtManager.Stop(ctx)
	ls, err := golegs.NewSubscriber(h, nil, lnkSys, PubSubTopic, nil,
		golegs.AllowPeer(reg.Authorized), golegs.DtManager(dtManager, gs))
	if err != nil {
		return nil, nil, err
	}

	if c.options.RateLimit.Enable {
		gs.RegisterOutgoingRequestHook(c.rateLimitHook())
	}
	dtManager.SubscribeToEvents(onDataTransferComplete)

	return ls, gs, nil
}

func (c *Core) Close() error {
	// Close leg transport.
	err := c.LS.Close()

	c.cancelSyncFn()
	<-c.watchDone
	c.waitForPendingSyncs.Wait()

	return err
}

func (c *Core) autoSync() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for provInfo := range c.reg.SyncChan() {
		c.waitForPendingSyncs.Add(1)
		go func(pubID, provID peer.ID, pubAddr multiaddr.Multiaddr) {
			defer c.waitForPendingSyncs.Done()

			log := logger.With("publisher", pubID, "provider", provID, "addr", pubAddr)
			log.Info("Auto-syncing the latest meta-data with publisher")

			_, err := c.LS.Sync(ctx, pubID, cid.Undef, nil, pubAddr)
			if err != nil {
				log.Errorw("Failed to auto-sync with publisher", "err", err)
				return
			}
		}(provInfo.Publisher, provInfo.AddrInfo.ID, provInfo.AddrInfo.Addrs[0])
	}
}

// restoreLatestSync reads the latest sync for each previously synced provider,
// from the datastore, and sets this in the Subscriber.
func (c *Core) restoreLatestSync() error {
	// Load all pins from the datastore.
	q := query.Query{
		Prefix: SyncPrefix,
	}
	results, err := c.DS.Query(context.Background(), q)
	if err != nil {
		return err
	}
	defer results.Close()

	var count int
	for r := range results.Next() {
		if r.Error != nil {
			return fmt.Errorf("cannot read latest syncs: %w", r.Error)
		}
		ent := r.Entry
		_, lastCid, err := cid.CidFromBytes(ent.Value)
		if err != nil {
			logger.Errorw("Failed to decode latest sync CID", "err", err)
			continue
		}
		if lastCid == cid.Undef {
			continue
		}
		peerID, err := peer.Decode(strings.TrimPrefix(ent.Key, SyncPrefix))
		if err != nil {
			logger.Errorw("Failed to decode peer ID of latest sync", "err", err)
			continue
		}

		err = c.LS.SetLatestSync(peerID, lastCid)
		if err != nil {
			logger.Errorw("Failed to set latest sync", "err", err, "peer", peerID)
			continue
		}
		logger.Debugw("Set latest sync", "provider", peerID, "cid", lastCid)
		count++
	}
	logger.Infow("Loaded latest sync for providers", "count", count)
	return nil
}

func (c *Core) SetRatelimiter(rl *policy.Limiter) {
	c.rateLimiter = rl
}

// watchSyncFinished reads legs.SyncFinished events and records the latest sync
// for the peer that was synced.
func (c *Core) watchSyncFinished(onSyncFin <-chan golegs.SyncFinished) {
	for syncFin := range onSyncFin {
		if _, err := c.PS.Get(context.Background(), syncFin.Cid); err != nil {
			// skip if data is not stored
			continue
		}

		exist, _ := c.checkCidCached(syncFin.Cid)
		if exist {
			// seen cid before, skip
			continue
		}

		metrics.Counter(context.Background(), metrics.ProviderNotificationCount, syncFin.PeerID.String(), 1)()

		// Persist the latest sync
		err := c.DS.Put(context.Background(), datastore.NewKey(SyncPrefix+syncFin.PeerID.String()), syncFin.Cid.Bytes())
		if err != nil {
			logger.Errorw("Error persisting latest sync", "err", err, "peer", syncFin.PeerID)
			continue
		}
		logger.Debugw("Persisted latest sync", "peer", syncFin.PeerID, "cid", syncFin.Cid)
	}
	close(c.watchDone)
}

func (c *Core) SendRecvMeta(mcid cid.Cid, mpeer peer.ID) {
	ctx, cncl := context.WithTimeout(context.Background(), time.Minute*10)
	defer cncl()
	select {
	case c.recvMetaCh <- &metadata.MetaRecord{
		Cid:        mcid,
		ProviderID: mpeer,
		Time:       uint64(time.Now().Unix()),
	}:
	case _ = <-ctx.Done():
		logger.Errorf("failed to send metadata(cid: %s peerid: %s) to metamanager, timeout", mcid.String(), mpeer.String())
	}

}

func (c *Core) checkCidCached(mcid cid.Cid) (bool, error) {
	err := c.CS.View(func(txn *badger.Txn) error {
		_, err := txn.Get(mcid.Bytes())
		return err
	})
	if err == nil {
		logger.Debugf("cid %s exists, ignore", mcid.String())
		return true, nil
	} else if err != badger.ErrKeyNotFound {
		logger.Warnf("an error occured when get cid %s from the cache: %v",
			mcid.String(), err)
		return false, err
	}
	// Save viewed cid
	err = c.CS.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(mcid.Bytes(), []byte("")).WithTTL(c.backupGenInterval)
		return txn.SetEntry(e)
	})
	if err != nil {
		logger.Warnf("cache cid %s failed, error: %v", mcid.String(), err)
		return false, err
	}
	logger.Debugf("Cached latest sync cid: %s", mcid.String())
	return false, nil
}
