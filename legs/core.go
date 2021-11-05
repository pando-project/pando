package legs

import (
	"Pando/metadata"
	"context"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/gammazero/keymutex"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

var log = logging.Logger("core")

const (
	// prefix used to track latest sync in datastore.
	syncPrefix  = "/sync/"
	admapPrefix = "/admap/"

	PUBSUBTOPIC = "PandoPubSub"
)

type LegsCore struct {
	Host           *host.Host
	DS             datastore.Batching
	BS             blockstore.Blockstore
	lms            golegs.LegMultiSubscriber
	subs           map[peer.ID]*subscriber
	sublk          *keymutex.KeyMutex
	receivedMetaCh chan<- *metadata.MetaRecord
}

// subscriber datastructure for a peer.
type subscriber struct {
	peerID  peer.ID
	ls      golegs.LegSubscriber
	watcher <-chan cid.Cid
	cncl    context.CancelFunc
}

func NewLegsCore(ctx context.Context, host *host.Host, ds datastore.Batching, bs blockstore.Blockstore, outMetaCh chan<- *metadata.MetaRecord) (*LegsCore, error) {

	lnkSys := MkLinkSystem(bs)

	lms, err := golegs.NewMultiSubscriber(ctx, *host, ds, lnkSys, PUBSUBTOPIC)
	if err != nil {
		return nil, err
	}

	lc := &LegsCore{
		Host:           host,
		DS:             ds,
		BS:             bs,
		lms:            lms,
		subs:           make(map[peer.ID]*subscriber),
		sublk:          keymutex.New(0),
		receivedMetaCh: outMetaCh,
	}

	lms.GraphSync().RegisterIncomingBlockHook(lc.storageHook())
	log.Debugf("LegCore started and all hooks and linksystem registered")

	return lc, nil
}

func (l *LegsCore) Subscribe(ctx context.Context, peerID peer.ID) error {
	log.Infow("Subscribing to advertisement pub-sub channel", "host_id", peerID)
	sctx, cancel := context.WithCancel(ctx)
	sub, err := l.newPeerSubscriber(sctx, peerID)
	if err != nil {
		log.Errorf("Error getting a subscriber instance for provider: %s", err)
		cancel()
		return err
	}

	// If already subscribed do nothing.
	if sub.watcher != nil {
		log.Infow("Already subscribed to provider", "id", peerID)
		cancel()
		return nil
	}

	var cncl context.CancelFunc
	sub.watcher, cncl = sub.ls.OnChange()
	sub.cncl = cancelFunc(cncl, cancel)

	// Listen updates persist latestSync when sync is done.
	go l.listenSubUpdates(sub)
	return nil
}

// Unsubscribe to stop listening to advertisement from a specific provider.
func (i *LegsCore) Unsubscribe(ctx context.Context, peerID peer.ID) error {
	log.Debugf("Unsubscribing from provider %s", peerID)
	i.sublk.Lock(string(peerID))
	defer i.sublk.Unlock(string(peerID))
	// Check if subscriber exists.
	sub, ok := i.subs[peerID]
	if !ok {
		log.Infof("Not subscribed to provider %s. Nothing to do", peerID)
		// If not we have nothing to do.
		return nil
	}
	// Close subscriber
	sub.ls.Close()
	// Check if we are subscribed
	if sub.cncl != nil {
		// If yes, run cancel
		sub.cncl()
	}
	// Delete from map
	delete(i.subs, peerID)
	log.Infof("Unsubscribed from provider %s successfully", peerID)

	return nil
}

// Creates a new subscriber for a peer according to its latest sync.
func (l *LegsCore) newPeerSubscriber(ctx context.Context, peerID peer.ID) (*subscriber, error) {
	l.sublk.Lock(string(peerID))
	defer l.sublk.Unlock(string(peerID))
	sub, ok := l.subs[peerID]
	// If there is already a subscriber for the peer, do nothing.
	if ok {
		return sub, nil
	}

	// See if we already synced with this peer.
	c, err := l.getLatestSync(peerID)
	if err != nil {
		return nil, err
	}

	// If not synced start a brand new subscriber
	var ls golegs.LegSubscriber
	if c == cid.Undef {
		ls, err = l.lms.NewSubscriber(golegs.FilterPeerPolicy(peerID))
	} else {
		// If yes, start a partially synced subscriber.
		ls, err = l.lms.NewSubscriberPartiallySynced(golegs.FilterPeerPolicy(peerID), c)
	}
	if err != nil {
		return nil, err
	}
	sub = &subscriber{
		peerID: peerID,
		ls:     ls,
	}
	l.subs[peerID] = sub
	return sub, nil
}

// Get the latest cid synced for the peer.
func (l *LegsCore) getLatestSync(peerID peer.ID) (cid.Cid, error) {
	b, err := l.DS.Get(datastore.NewKey(syncPrefix + peerID.String()))
	if err != nil {
		if err == datastore.ErrNotFound {
			return cid.Undef, nil
		}
		return cid.Undef, err
	}
	_, c, err := cid.CidFromBytes(b)
	return c, err
}

// cancelfunc for subscribers. Combines context cancel and LegSubscriber
// cancel function.
func cancelFunc(c1, c2 context.CancelFunc) context.CancelFunc {
	return func() {
		c1()
		c2()
	}
}

func (l *LegsCore) listenSubUpdates(sub *subscriber) {
	for c := range sub.watcher {
		// Persist the latest sync
		if err := l.putLatestSync(sub.peerID, c); err != nil {
			log.Errorf("Error persisting latest sync: %s", err)
		}

		//todo debug
		v, e := l.BS.Get(c)
		if e != nil {
			log.Errorf("not received the block for: %s", c.String())
		}
		log.Debugf("block value: %s", string(v.RawData()))

		l.receivedMetaCh <- &metadata.MetaRecord{
			Cid:        c,
			ProviderID: sub.peerID,
			Time:       uint64(time.Now().Unix()),
		}
	}
}

// Tracks latest sync for a specific peer.
func (l *LegsCore) putLatestSync(peerID peer.ID, c cid.Cid) error {
	// Do not save if empty CIDs are received. Closing the channel
	// may lead to receiving empty CIDs.
	if c == cid.Undef {
		return nil
	}

	return l.DS.Put(datastore.NewKey(syncPrefix+peerID.String()), c.Bytes())
}
