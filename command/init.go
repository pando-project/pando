package command

import (
	"Pando/config"
	"fmt"
	"os"

	//"github.com/filecoin-project/storetheindex/config"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var InitCmd = &cli.Command{
	Name:   "init",
	Usage:  "Initialize indexer node config file and identity",
	Flags:  initFlags,
	Action: initCommand,
}

func initCommand(cctx *cli.Context) error {
	// Check that the config root exists and it writable.
	configRoot, err := config.PathRoot()
	if err != nil {
		return err
	}

	fmt.Println("Initializing indexer node at", configRoot)

	if err = checkWritable(configRoot); err != nil {
		return err
	}

	configFile, err := config.Filename(configRoot)
	if err != nil {
		return err
	}

	if fileExists(configFile) {
		return config.ErrInitialized
	}

	var cfg *config.Config
	speedTest := cctx.Bool("speedtest")
	if !speedTest {
		cfg, err = config.Init(os.Stderr, config.DisableTestSpeed())
	} else {
		cfg, err = config.Init(os.Stderr)
	}
	if err != nil {
		return err
	}

	adminAddr := cctx.String("listen-admin")
	if adminAddr != "" {
		_, err := multiaddr.NewMultiaddr(adminAddr)
		if err != nil {
			return fmt.Errorf("bad listen-admin: %s", err)
		}
		cfg.Addresses.AdminServer = adminAddr
	}

	pandoAddr := cctx.String("listen-pando")
	if pandoAddr != "" {
		_, err := multiaddr.NewMultiaddr(pandoAddr)
		if err != nil {
			return fmt.Errorf("bad listen-pando: %s", err)
		}
		cfg.Addresses.PandoServer = pandoAddr
	}

	return cfg.Save(configFile)
}
