package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/pkg/registry/discovery"
	"github.com/kenlabs/pando/pkg/registry/internal/syserr"
	"github.com/kenlabs/pando/pkg/registry/policy"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	// providerKeyPath is where provider info is stored in to indexer repo
	providerKeyPath = "/registry/pinfo"
)

var log = logging.Logger("registry")

// Registry stores information about discovered providers
type Registry struct {
	actions   chan func()
	closed    chan struct{}
	closeOnce sync.Once
	dstore    datastore.Datastore
	providers map[peer.ID]*ProviderInfo
	sequences *sequences

	//core legs_interface.PandoCore

	discoverer   discovery.Discoverer
	discoWait    sync.WaitGroup
	discoTimes   map[string]time.Time
	policy       *policy.Policy
	accountLevel *option.AccountLevel

	discoveryTimeout time.Duration
	pollInterval     time.Duration
	rediscoverWait   time.Duration

	periodicTimer *time.Timer
}

// ProviderInfo is an immutable data sturcture that holds information about a
// provider.  A ProviderInfo instance is never modified, but rather a new one
// is created to update its contents.  This means existing references remain
// valid.
type ProviderInfo struct {
	// AddrInfo contains an account.ID and set of Multiaddr addresses.
	AddrInfo peer.AddrInfo
	// DiscoveryAddr is the address that is used for discovery of the provider.
	DiscoveryAddr string
	// AccountLevel is the level according to the filecoin miner account balance
	AccountLevel int

	LastBackupMeta cid.Cid
}

func (p *ProviderInfo) dsKey() datastore.Key {
	return datastore.NewKey(path.Join(providerKeyPath, p.AddrInfo.ID.String()))
}

// NewRegistry creates a new provider registry, giving it provider policy
// configuration, a datastore to persist provider data, and a Discoverer
// interface.
//
// TODO: It is probably necessary to have multiple discoverer interfaces
func NewRegistry(ctx context.Context, cfg *option.Discovery, cfglevel *option.AccountLevel, dstore datastore.Datastore, disco discovery.Discoverer) (*Registry, error) {
	if cfg == nil || cfglevel == nil {
		return nil, fmt.Errorf("nil config")
	}

	// Create policy from config
	discoPolicy, err := policy.New(cfg.Policy)
	if err != nil {
		return nil, err
	}

	r := &Registry{
		actions:   make(chan func()),
		closed:    make(chan struct{}),
		policy:    discoPolicy,
		providers: map[peer.ID]*ProviderInfo{},
		sequences: newSequences(0),

		pollInterval:     time.Duration(cfg.PollIntervalInDurationFormat()),
		rediscoverWait:   time.Duration(cfg.RediscoverWaitInDurationFormat()),
		discoveryTimeout: time.Duration(cfg.TimeoutInDurationFormat()),

		accountLevel: cfglevel,

		discoverer: disco,
		discoTimes: map[string]time.Time{},

		dstore: dstore,
		//core:   core,
	}

	count, err := r.loadPersistedProviders(ctx)
	if err != nil {
		return nil, err
	}
	log.Infow("loaded providers into registry", "count", count)

	r.periodicTimer = time.AfterFunc(r.pollInterval/2, func() {
		r.cleanup()
		r.periodicTimer.Reset(r.pollInterval / 2)
	})

	go r.run()
	return r, nil
}

// Close waits for any pending discoverer to finish and then stops the registry
func (r *Registry) Close() error {
	var err error
	r.closeOnce.Do(func() {
		r.periodicTimer.Stop()
		// Wait for any pending discoveries to complete, then stop the main run
		// goroutine
		r.discoWait.Wait()
		close(r.actions)

		if r.dstore != nil {
			err = r.dstore.Close()
		}
	})
	<-r.closed
	return err
}

// run executs functions that need to be executed on the same goroutine
//
// Running actions here is a substitute for mutex-locking the sections of code
// run as an action and allows the caller to decide whether or not to wait for
// the code to finish running.
//
// All functions named using the prefix "sync" must be run on this goroutine.
func (r *Registry) run() {
	defer close(r.closed)

	for action := range r.actions {
		action()
	}
}

// Register is used to directly register a provider, bypassing discovery and
// adding discovered data directly to the registry.
func (r *Registry) Register(ctx context.Context, info *ProviderInfo) error {
	// If provider is not allowed, then ignore request
	if !r.policy.Allowed(info.AddrInfo.ID) {
		return syserr.New(ErrNotAllowed, http.StatusForbidden)
	}

	// If provider is trusted, register immediately without verification
	if !r.policy.Trusted(info.AddrInfo.ID) {
		return syserr.New(ErrNotTrusted, http.StatusUnauthorized)
	}
	// info should not contain the weight before evaluating
	if info.AccountLevel != 0 {
		return syserr.New(ErrWrongWeight, http.StatusBadRequest)
	}

	// If provider have miner account, discover it
	if info.DiscoveryAddr != "" {
		log.Infow("found miner account, start discovering")
		discoveredData, err := r.discoverer.Discover(context.Background(), info.AddrInfo.ID, info.DiscoveryAddr)
		if err != nil {
			log.Infof("discovering failed: %s", err.Error())
			return fmt.Errorf("discovering failed: %s", err.Error())
		}
		info.AccountLevel, err = r.getAccountLevel(discoveredData.Balance)
		if err != nil {
			log.Warnf("falied to get the account level. %s", err.Error())
			return fmt.Errorf("falied to get the account level. %s", err.Error())
		}
		log.Debugf("discovering successed, peerID: %s, account balance: %s", info.AddrInfo.ID.String(), discoveredData.Balance.String())
	}

	errCh := make(chan error, 1)
	r.actions <- func() {
		errCh <- r.syncRegister(ctx, info)
	}

	err := <-errCh
	if err != nil {
		return err
	}

	log.Infow("registered provider", "id", info.AddrInfo.ID, "addrs", info.AddrInfo.Addrs)

	return nil
}

// IsRegistered checks if the provider is in the registry
func (r *Registry) IsRegistered(providerID peer.ID) bool {
	done := make(chan struct{})
	var found bool
	r.actions <- func() {
		_, found = r.providers[providerID]
		close(done)
	}
	<-done
	return found
}

// IsTrusted checks if the provider is in the white list
func (r *Registry) IsTrusted(providerID peer.ID) bool {
	return r.policy.Trusted(providerID)
}

// ProviderInfoByAddr finds a registered provider using its discovery address
func (r *Registry) ProviderInfoByAddr(discoAddr string) *ProviderInfo {
	infoChan := make(chan *ProviderInfo)
	r.actions <- func() {
		// TODO: consider adding a map of discoAddr->providerID
		for _, info := range r.providers {
			if info.DiscoveryAddr == discoAddr {
				infoChan <- info
				break
			}
		}
		close(infoChan)
	}

	return <-infoChan
}

// ProviderInfo returns information for a registered provider
func (r *Registry) ProviderInfo(providerID peer.ID) []*ProviderInfo {
	if providerID == "" {
		return r.AllProviderInfo()
	}

	infoChan := make(chan *ProviderInfo)
	r.actions <- func() {
		info, ok := r.providers[providerID]
		if ok {
			infoChan <- info
		}
		close(infoChan)
	}

	return []*ProviderInfo{<-infoChan}
}

// AllProviderInfo returns information for all registered providers
func (r *Registry) AllProviderInfo() []*ProviderInfo {
	var infos []*ProviderInfo
	done := make(chan struct{})
	r.actions <- func() {
		infos = make([]*ProviderInfo, len(r.providers))
		var i int
		for _, info := range r.providers {
			infos[i] = info
			i++
		}
		close(done)
	}
	<-done
	return infos
}

func (r *Registry) CheckSequence(peerID peer.ID, seq uint64) error {
	return r.sequences.check(peerID, seq)
}

func (r *Registry) syncRegister(ctx context.Context, info *ProviderInfo) error {
	r.providers[info.AddrInfo.ID] = info
	err := r.syncPersistProvider(ctx, info)
	if err != nil {
		err = fmt.Errorf("could not persist provider: %s", err)
		return syserr.New(err, http.StatusInternalServerError)
	}
	return nil
}

func (r *Registry) syncPersistProvider(ctx context.Context, info *ProviderInfo) error {
	if r.dstore == nil {
		// todo  why not return error?
		//return fmt.Errorf("nil datastore")
		return nil
	}
	value, err := json.Marshal(info)
	if err != nil {
		return err
	}

	dsKey := info.dsKey()
	if err = r.dstore.Put(ctx, dsKey, value); err != nil {
		return err
	}
	if err = r.dstore.Sync(ctx, dsKey); err != nil {
		return fmt.Errorf("cannot sync provider info: %s", err)
	}
	return nil
}

func (r *Registry) loadPersistedProviders(ctx context.Context) (int, error) {
	if r.dstore == nil {
		return 0, nil
	}

	// Load all providers from the datastore.
	q := query.Query{
		Prefix: providerKeyPath,
	}
	results, err := r.dstore.Query(ctx, q)
	if err != nil {
		return 0, err
	}
	defer results.Close()

	var count int
	for result := range results.Next() {
		if result.Error != nil {
			return 0, fmt.Errorf("cannot read provider data: %v", result.Error)
		}
		ent := result.Entry

		peerID, err := peer.Decode(path.Base(ent.Key))
		if err != nil {
			return 0, fmt.Errorf("cannot decode provider ID: %s", err)
		}

		pinfo := new(ProviderInfo)
		err = json.Unmarshal(ent.Value, pinfo)
		if err != nil {
			return 0, err
		}

		r.providers[peerID] = pinfo
		count++
	}
	return count, nil
}

// Check if the peer is trusted by policy, or if it has been previously
// verified and registered as a provider.
func (r *Registry) Authorized(peerID peer.ID) (bool, error) {
	allowed, trusted := r.policy.Check(peerID)

	if !allowed {
		return false, nil
	}

	// Peer is allowed but not trusted, see if it is a registered provider.
	if !trusted {
		regOk := make(chan bool)
		r.actions <- func() {
			_, ok := r.providers[peerID]
			regOk <- ok
		}
		return <-regOk, nil
	}

	return true, nil
}

// RegisterOrUpdate attempts to register an unregistered provider, or updates
// the addresses and latest meta data of an already registered provider.
func (r *Registry) RegisterOrUpdate(ctx context.Context, providerID peer.ID, metaCid cid.Cid) error {
	// Check that the provider has been discovered and validated
	info := r.ProviderInfo(providerID)[0]
	if info != nil {
		info = &ProviderInfo{
			AddrInfo: peer.AddrInfo{
				ID:    providerID,
				Addrs: info.AddrInfo.Addrs,
			},
			DiscoveryAddr:  info.DiscoveryAddr,
			LastBackupMeta: info.LastBackupMeta,
		}
	} else {
		info = &ProviderInfo{
			AddrInfo: peer.AddrInfo{
				ID: providerID,
			},
		}
	}

	if metaCid != cid.Undef {
		info.LastBackupMeta = metaCid
	}

	return r.Register(ctx, info)
}

func (r *Registry) cleanup() {
	r.discoWait.Add(1)
	r.sequences.retire()
	r.actions <- func() {
		now := time.Now()
		for id, completed := range r.discoTimes {
			if completed.IsZero() {
				continue
			}
			if r.rediscoverWait != 0 && now.Sub(completed) < r.rediscoverWait {
				continue
			}
			delete(r.discoTimes, id)
		}
		if len(r.discoTimes) == 0 {
			r.discoTimes = make(map[string]time.Time)
		}
	}
	r.discoWait.Done()
}
