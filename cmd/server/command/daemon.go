package command

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	mutexDataStoreFactory "github.com/ipfs/go-datastore/sync"
	dataStoreFactory "github.com/ipfs/go-ds-leveldb"
	blockStoreFactory "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/lotus"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/policy"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/kenlabs/pando/pkg/statetree"
	"github.com/kenlabs/pando/pkg/statetree/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p-core/host"
	"github.com/spf13/cobra"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kenlabs/pando/pkg/system"
)

var log = logging.Logger("pando")

func DaemonCmd() *cobra.Command {
	const failedError = "run daemon failed: \n\t%v\n"
	return &cobra.Command{
		Use:   "daemon",
		Short: "start pando server",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := setLoglevel()
			if err != nil {
				return fmt.Errorf(failedError, err)
			}

			storeInstance, err := initStoreInstance()
			if err != nil {
				return fmt.Errorf(failedError, err)
			}
			defer storeInstance.CacheStore.Close()

			privateKey, err := Opt.Identity.DecodePrivateKey()
			if err != nil {
				return fmt.Errorf(failedError, err)
			}

			p2pHost, err := initP2PHost(privateKey)
			if err != nil {
				return fmt.Errorf(failedError, err)
			}

			c, err := initCore(storeInstance, p2pHost)
			if err != nil {
				return fmt.Errorf(failedError, err)
			}

			server, err := api.NewAPIServer(Opt, c)
			if err != nil {
				return fmt.Errorf(failedError, err)
			}

			server.MustStartAllServers()

			quit := make(chan os.Signal)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit
			fmt.Println("Shutting down servers...")
			err = server.StopAllServers()
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func setLoglevel() error {
	// Supported LogLevel are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and
	// their lower-case forms.
	logLevel, err := logging.LevelFromString(Opt.LogLevel)
	if err != nil {
		return err
	}
	logging.SetAllLoggers(logLevel)
	//err = logging.SetLogLevel("addrutil", "warn")
	//if err != nil {
	//	return err
	//}
	err = logging.SetLogLevel("basichost", "warn")
	if err != nil {
		return err
	}
	err = logging.SetLogLevel("meta-manager", "warn")
	if err != nil {
		return err
	}

	return nil
}

func initStoreInstance() (*core.StoreInstance, error) {
	if Opt.DataStore.Type != "levelds" {
		return nil, fmt.Errorf("only levelds datastore type supported")
	}

	dataStoreDir := filepath.Join(Opt.PandoRoot, Opt.DataStore.Dir)
	dataStoreDirExists, err := system.IsDirExists(dataStoreDir)
	if err != nil {
		return nil, err
	}
	if !dataStoreDirExists {
		err := os.MkdirAll(dataStoreDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	writable, err := system.IsDirWritable(dataStoreDir)
	if err != nil {
		return nil, err
	}
	if !writable {
		return nil, err
	}

	dataStore, err := dataStoreFactory.NewDatastore(dataStoreDir, nil)
	if err != nil {
		return nil, err
	}
	mutexDataStore := mutexDataStoreFactory.MutexWrap(dataStore)
	blockStore := blockStoreFactory.NewBlockstore(mutexDataStore)

	cacheStoreDir := filepath.Join(Opt.PandoRoot, Opt.CacheStore.Dir)
	cacheStore, err := badger.Open(badger.DefaultOptions(cacheStoreDir))

	return &core.StoreInstance{
		DataStore:      dataStore,
		CacheStore:     cacheStore,
		MutexDataStore: mutexDataStore,
		BlockStore:     blockStore,
	}, nil
}

func initP2PHost(privateKey crypto.PrivKey) (libp2pHost.Host, error) {
	var p2pHost libp2pHost.Host
	var err error
	if !Opt.ServerAddress.DisableP2P {
		log.Info("initializing libp2p host...")
		p2pHost, err = libp2p.New(
			libp2p.ListenAddrStrings(Opt.ServerAddress.P2PAddress),
			libp2p.Identity(privateKey),
		)
		if err != nil {
			return nil, err
		}
		log.Debugf("multiaddr is: %s", p2pHost.Addrs())
		log.Debugf("peerID is: %s", p2pHost.ID())
	} else {
		log.Info("libp2p host disabled")
	}

	return p2pHost, nil
}

func initCore(storeInstance *core.StoreInstance, p2pHost libp2pHost.Host) (*core.Core, error) {
	c := &core.Core{}
	var err error

	c.StoreInstance = storeInstance
	linkSystem := legs.MkLinkSystem(c.StoreInstance.BlockStore, nil)
	c.LinkSystem = &linkSystem

	var lotusDiscoverer *lotus.Discoverer
	if Opt.Discovery.LotusGateway != "" {
		log.Infow("discovery using lotus", "gateway", Opt.Discovery.LotusGateway)
		// Create lotus client
		c.LotusDiscover, err = lotus.NewDiscoverer(Opt.Discovery.LotusGateway)
		if err != nil {
			return nil, fmt.Errorf("cannot create lotus client: %v", err)
		}
	}

	c.Registry, err = registry.NewRegistry(context.Background(), &Opt.Discovery, &Opt.AccountLevel,
		storeInstance.DataStore, lotusDiscoverer)
	if err != nil {
		return nil, fmt.Errorf("cannot create provider registryInstance: %v", err)
	}

	c.MetaManager, err = metadata.New(context.Background(),
		storeInstance.MutexDataStore,
		storeInstance.BlockStore,
		c.LinkSystem,
		c.Registry,
		&Opt.Backup)
	if err != nil {
		return nil, err
	}

	info := new(types.ExtraInfo)
	for _, addr := range p2pHost.Addrs() {
		info.MultiAddresses += addr.String() + " "
	}
	info.PeerID = p2pHost.ID().String()

	c.StateTree, err = statetree.New(context.Background(),
		storeInstance.MutexDataStore,
		storeInstance.BlockStore,
		c.MetaManager.GetUpdateOut(),
		info,
	)
	if err != nil {
		return nil, err
	}

	backupGenInterval, err := time.ParseDuration(Opt.Backup.BackupGenInterval)
	if err != nil {
		return nil, err
	}
	c.LegsCore, err = legs.NewLegsCore(context.Background(),
		p2pHost,
		storeInstance.MutexDataStore,
		storeInstance.CacheStore,
		storeInstance.BlockStore,
		c.MetaManager.GetMetaInCh(),
		backupGenInterval,
		nil,
		c.Registry,
		Opt,
	)

	tokenRate := math.Ceil((0.8 * float64(Opt.RateLimit.Bandwidth)) / Opt.RateLimit.SingleDAGSize)
	rateConfig := &policy.LimiterConfig{
		TotalRate:     tokenRate,
		TotalBurst:    int(math.Ceil(tokenRate)),
		BaseTokenRate: tokenRate,
		Registry:      c.Registry,
	}
	rateLimiter, err := policy.NewLimiter(*rateConfig)
	if err != nil {
		return nil, err
	}

	c.LegsCore.SetRatelimiter(rateLimiter)

	return c, nil
}
