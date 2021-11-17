package command

import (
	v0client "Pando/api/v0/admin/client/http"
	"Pando/config"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
)

var RegisterCmd = &cli.Command{
	Name:   "register",
	Usage:  "Register provider information with an indexer that trusts the provider",
	Flags:  registerFlags,
	Action: registerCommand,
}

func registerCommand(cctx *cli.Context) error {
	cfg, err := config.Load(cctx.String("config"))
	if err != nil {
		if err == config.ErrNotInitialized {
			err = errors.New("config file not found")
		}
		return fmt.Errorf("cannot load config file: %w", err)
	}

	peerID, privKey, err := cfg.Identity.Decode()
	if err != nil {
		return err
	}

	client, err := v0client.New(cctx.String("pando"))
	if err != nil {
		return err
	}

	err = client.Register(cctx.Context, peerID, privKey, cctx.StringSlice("provider-addr"), cctx.String("miner"))
	if err != nil {
		return fmt.Errorf("failed to register providers: %s", err)
	}

	fmt.Println("Registered provider", cfg.Identity.PeerID, "at pando", cctx.String("pando"))
	return nil
}