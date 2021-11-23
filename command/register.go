package command

import (
	v0client "Pando/api/v0/admin/client/http"
	"Pando/config"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
)

var RegisterCmd = &cli.Command{
	Name:   "register",
	Usage:  "Register provider information with an indexer that trusts the provider",
	Flags:  registerFlags,
	Action: registerCommand,
}

func registerCommand(cctx *cli.Context) error {
	f, err := os.Open(cctx.String("config"))
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg config.Identity
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return err
	}

	if cfg.PrivKey == "" || cfg.PeerID == "" {
		return fmt.Errorf("valid config")
	}
	peerID, privKey, err := cfg.Decode()
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

	fmt.Println("Registered provider", cfg.PeerID, "at pando", cctx.String("pando"))
	return nil
}
