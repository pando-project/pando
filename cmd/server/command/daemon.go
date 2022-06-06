package command

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	mutexDataStoreFactory "github.com/ipfs/go-datastore/sync"
	dataStoreFactory "github.com/ipfs/go-ds-leveldb"
	logging "github.com/ipfs/go-log/v2"
	PandoStore "github.com/kenlabs/PandoStore/pkg"
	"github.com/kenlabs/PandoStore/pkg/config"
	"github.com/kenlabs/PandoStore/pkg/migrate"
	"github.com/kenlabs/PandoStore/pkg/store"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/api/v1/server"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/lotus"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/policy"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/kenlabs/pando/pkg/util/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	libp2pHost "github.com/libp2p/go-libp2p-core/host"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kenlabs/pando/pkg/system"
)

var logger = log.NewSubsystemLogger()

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

			mongoClient, err := connectMetaCache(Opt.MetaCache.Type, Opt.MetaCache.ConnectionURI)
			if err != nil {
				return fmt.Errorf(failedError, err)
			}
			Opt.MetaCache.Client = mongoClient

			storeInstance, err := initStoreInstance()
			if err != nil {
				return fmt.Errorf(failedError, err)
			}
			defer storeInstance.CacheStore.Close()

			_, privateKey, err := Opt.Identity.Decode()
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

			server, err := server.NewAPIServer(Opt, c)
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
	err = logging.SetLogLevel("basichost", "warn")
	if err != nil {
		return err
	}
	//err = logging.SetLogLevel("meta-manager", "warn")
	//if err != nil {
	//	return err
	//}

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

	version, err := PandoStore.CheckVersion(dataStoreDir)
	if err != nil {
		return nil, err
	}
	if version != PandoStore.CurrentVersion {
		err = migrate.Migrate(version, PandoStore.CurrentVersion, dataStoreDir, false)
		if err != nil {
			return nil, err
		}
	}

	dataStore, err := dataStoreFactory.NewDatastore(dataStoreDir, nil)
	if err != nil {
		return nil, err
	}
	mutexDataStore := mutexDataStoreFactory.MutexWrap(dataStore)
	//blockStore := blockStoreFactory.NewBlockstore(mutexDataStore)

	cacheStoreDir := filepath.Join(Opt.PandoRoot, Opt.CacheStore.Dir)
	cacheStore, err := badger.Open(badger.DefaultOptions(cacheStoreDir))
	if err != nil {
		return nil, err
	}

	pandoStore, err := store.NewStoreFromDatastore(context.Background(), mutexDataStore, &config.StoreConfig{
		SnapShotInterval: Opt.DataStore.SnapShotInterval,
	})
	if err != nil {
		return nil, err
	}

	return &core.StoreInstance{
		//DataStore:      dataStore,
		CacheStore:     cacheStore,
		MutexDataStore: mutexDataStore,
		PandoStore:     pandoStore,
		MetadataCache:  Opt.MetaCache.Client,
	}, nil
}

func initP2PHost(privateKey crypto.PrivKey) (libp2pHost.Host, error) {
	var p2pHost libp2pHost.Host
	var err error

	logger.Info("initializing libp2p host...")
	if Opt.ServerAddress.P2PAddress == "" {
		return nil, fmt.Errorf("valid p2p address")
	}
	p2pHost, err = libp2p.New(
		libp2p.ListenAddrStrings(Opt.ServerAddress.P2PAddress),
		libp2p.Identity(privateKey),
	)
	if err != nil {
		return nil, err
	}
	logger.Debugf("multiaddr is: %s", p2pHost.Addrs())
	logger.Debugf("peerID is: %s", p2pHost.ID())

	return p2pHost, nil
}

func initCore(storeInstance *core.StoreInstance, p2pHost libp2pHost.Host) (*core.Core, error) {
	c := &core.Core{}
	var err error

	c.StoreInstance = storeInstance
	linkSystem := legs.MkLinkSystem(c.StoreInstance.PandoStore, nil, nil)
	c.LinkSystem = &linkSystem

	var lotusDiscoverer *lotus.Discoverer
	if Opt.Discovery.LotusGateway != "" {
		logger.Infow("discovery using lotus", "gateway", Opt.Discovery.LotusGateway)
		// Create lotus client
		c.LotusDiscover, err = lotus.NewDiscoverer(Opt.Discovery.LotusGateway)
		if err != nil {
			return nil, fmt.Errorf("cannot create lotus client: %v", err)
		}
	}

	c.Registry, err = registry.NewRegistry(context.Background(), &Opt.Discovery, &Opt.AccountLevel,
		storeInstance.MutexDataStore, lotusDiscoverer)
	if err != nil {
		return nil, fmt.Errorf("cannot create provider registryInstance: %v", err)
	}

	c.MetaManager, err = metadata.New(context.Background(),
		storeInstance.MutexDataStore,
		c.LinkSystem,
		c.Registry,
		&Opt.Backup)
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
		storeInstance.PandoStore,
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

func connectMetaCache(storeType string, connectionURI string) (*mongo.Client, error) {
	switch storeType {
	case "mongodb":
		clientOptions := options.Client().ApplyURI(connectionURI)
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			return nil, err
		}
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("metadata store type: %s not supported", storeType)
	}

}
