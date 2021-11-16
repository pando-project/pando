package command

import (
	"Pando/config"
	"Pando/internal/lotus"
	"Pando/internal/registry"
	"Pando/legs"
	"Pando/metadata"
	httpadminserver "Pando/server/admin/http"
	graphserver "Pando/server/graph_sync/http"
	metaserver "Pando/server/metadata/http"
	"Pando/statetree"
	"Pando/statetree/types"
	"context"
	"errors"
	"fmt"
	dssync "github.com/ipfs/go-datastore/sync"
	leveldb "github.com/ipfs/go-ds-leveldb"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var log = logging.Logger("pando")

const shutdownTimeout = 5 * time.Second

var (
	ErrDaemonStart = errors.New("daemon did not start correctly")
	ErrDaemonStop  = errors.New("daemon did not stop correctly")
)

var DaemonCmd = &cli.Command{
	Name:   "daemon",
	Usage:  "start a pando daemon, accepting http requests",
	Flags:  nil,
	Action: daemonCommand,
}

func daemonCommand(cctx *cli.Context) error {
	err := logging.SetLogLevel("pando", "debug")
	if err != nil {
		return err
	}
	_ = logging.SetLogLevel("core", "debug")
	_ = logging.SetLogLevel("graphsync", "debug")
	_ = logging.SetLogLevel("meta-manager", "debug")
	_ = logging.SetLogLevel("state-tree", "debug")
	_ = logging.SetLogLevel("meta-server", "debug")
	_ = logging.SetLogLevel("admin", "debug")
	_ = logging.SetLogLevel("registry", "debug")

	cfg, err := config.Load("")
	if err != nil {
		if err == config.ErrNotInitialized {
			fmt.Fprintln(os.Stderr, "pando is not initialized")
			os.Exit(1)
		}
		return fmt.Errorf("cannot load config file: %w", err)
	}

	if cfg.Datastore.Type != "levelds" {
		return fmt.Errorf("only levelds datastore type supported, %q not supported", cfg.Datastore.Type)
	}

	// Create datastore
	dataStorePath, err := config.Path("", cfg.Datastore.Dir)
	if err != nil {
		return err
	}
	err = checkWritable(dataStorePath)
	if err != nil {
		return err
	}
	dstore, err := leveldb.NewDatastore(dataStorePath, nil)
	if err != nil {
		return err
	}
	mds := dssync.MutexWrap(dstore)
	bs := blockstore.NewBlockstore(mds)

	privKey, err := cfg.Identity.DecodePrivateKey("")
	p2pHost, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrStrings(cfg.Addresses.P2PAddr),
		libp2p.Identity(privKey),
	)
	if err != nil {
		return err
	}
	log.Debugf("multiaddr is: %s", p2pHost.Addrs())
	log.Debugf("peerID is: %s", p2pHost.ID())

	if err != nil {
		return err
	}
	metaManager, err := metadata.New(context.Background(), mds, bs)
	if err != nil {
		return err
	}
	info := new(types.ExtraInfo)
	info.MultiAddr = p2pHost.Addrs()[0].String()
	stateTree, err := statetree.New(context.Background(), mds, bs, metaManager.GetUpdateOut(), info)
	if err != nil {
		return err
	}
	legsCore, err := legs.NewLegsCore(context.Background(), &p2pHost, mds, bs, metaManager.GetMetaInCh())
	if err != nil {
		return err
	}

	var lotusDiscoverer *lotus.Discoverer
	if cfg.Discovery.LotusGateway != "" {
		log.Infow("discovery using lotus", "gateway", cfg.Discovery.LotusGateway)
		// Create lotus client
		lotusDiscoverer, err = lotus.NewDiscoverer(cfg.Discovery.LotusGateway)
		if err != nil {
			return fmt.Errorf("cannot create lotus client: %s", err)
		}
	}

	// Create registry
	registry, err := registry.NewRegistry(cfg.Discovery, dstore, lotusDiscoverer)
	if err != nil {
		return fmt.Errorf("cannot create provider registry: %s", err)
	}

	// http servers
	graphSyncServer, err := graphserver.New(cfg.Addresses.GraphSync, legsCore)
	if err != nil {
		return err
	}
	adminServer, err := httpadminserver.New(cfg.Addresses.Admin, registry)
	if err != nil {
		return err
	}
	metaDataServer, err := metaserver.New(cfg.Addresses.MetaData, cfg.Addresses.GraphQL, stateTree)
	if err != nil {
		return err
	}

	log.Info("Starting http servers")
	errChan := make(chan error, 1)
	go func() {
		errChan <- graphSyncServer.Start()
	}()
	go func() {
		errChan <- metaDataServer.Start()
	}()
	go func() {
		errChan <- adminServer.Start()
	}()

	var finalErr error
	select {
	case <-cctx.Done():
		// Command was canceled (ctrl-c)
	case err = <-errChan:
		log.Errorw("Failed to start server", "err", err)
		finalErr = ErrDaemonStart
	}

	log.Infow("Shutting down daemon")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	go func() {
		// Wait for context to be canceled.  If timeout, then exit with error.
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Timed out on shutdown, terminating...")
			os.Exit(-1)
		}
	}()

	if err = graphSyncServer.Shutdown(ctx); err != nil {
		log.Errorw("Error shutting down graphsync server", "err", err)
		finalErr = ErrDaemonStop
	}
	if err = metaDataServer.Shutdown(ctx); err != nil {
		log.Errorw("Error shutting down metadata server", "err", err)
		finalErr = ErrDaemonStop
	}
	if err = adminServer.Shutdown(ctx); err != nil {
		log.Errorw("Error shutting down admin server", "err", err)
		finalErr = ErrDaemonStop
	}

	cancel()

	log.Info("pando stopped")

	return finalErr
}
