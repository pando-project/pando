package command

import (
	v0client "Pando/api/v0/admin/client/http"
	"Pando/config"
	"encoding/json"
	"fmt"
	p2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
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
	privkeyStr := cctx.String("privkey")
	peeridStr := cctx.String("peerid")
	configPath := cctx.String("config")
	providerAddrStr := cctx.StringSlice("provider-addr")
	var err error

	for _, s := range providerAddrStr {
		_, err = multiaddr.NewMultiaddr(s)
		if err != nil {
			//return fmt.Errorf("invalid multiaddr: %s", s)
			//fmt.Println(s)
		}
	}

	if (privkeyStr == "" || peeridStr == "") && configPath == "" {
		return fmt.Errorf("please input private key and peerid or config file path")
	}

	var peerID peer.ID
	var privKey p2pcrypto.PrivKey
	if configPath == "" {
		peerID, privKey, err = config.Identity{
			PeerID:  peeridStr,
			PrivKey: privkeyStr,
		}.Decode()
		if err != nil {
			return err
		}
	} else {
		f, err := os.Open(cctx.String("config"))
		if err != nil {
			return err
		}
		defer f.Close()

		var cfg config.Identity
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return err
		}

		peerID, privKey, err = cfg.Decode()
		if err != nil {
			return err
		}
	}

	client, err := v0client.New(cctx.String("pando"))
	if err != nil {
		return err
	}

	err = client.Register(cctx.Context, peerID, privKey, providerAddrStr, cctx.String("miner"))
	if err != nil {
		return fmt.Errorf("failed to register providers: %s", err)
	}

	fmt.Println("Registered provider", peerID.String(), "at pando", cctx.String("pando"))
	return nil
}
