package command

import (
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
	"pando/pkg/register"
)

type providerInfo struct {
	peerID      string
	privateKey  string
	addresses   []string
	miner       string
	onlyEnvelop bool
}

var ProviderInfo *providerInfo

func RegisterCmd() *cobra.Command {
	ProviderInfo = &providerInfo{}

	registerCmd := &cobra.Command{
		Use:   "register",
		Short: "register a provider to pando server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ProviderInfo.validateFlags(); err != nil {
				return err
			}

			peerID, err := peer.Decode(ProviderInfo.peerID)
			if err != nil {
				return nil
			}

			privateKeyEncoded, err := base64.StdEncoding.DecodeString(ProviderInfo.privateKey)
			if err != nil {
				return err
			}
			privateKey, err := crypto.UnmarshalPrivateKey(privateKeyEncoded)
			if err != nil {
				return err
			}

			data, err := register.MakeRegisterRequest(peerID, privateKey, ProviderInfo.addresses, ProviderInfo.miner)

			if ProviderInfo.onlyEnvelop {
				fmt.Println(data)
				return nil
			}

			res, err := PandoClient.R().
				SetBody(data).
				SetHeader("Content-Type", "application/octet-stream").
				Post("/provider/register")
			if err != nil {
				return nil
			}
			if res.IsError() {
				return fmt.Errorf("response error: %v", res.Error())
			}

			fmt.Println("register success")

			return nil
		},
	}

	registerCmd.Flags().StringVar(&ProviderInfo.peerID, "peer-id", "",
		"peerID of provider, required")
	registerCmd.Flags().StringVar(&ProviderInfo.privateKey, "private-key", "",
		"private key of provider, required")
	registerCmd.Flags().StringSliceVar(&ProviderInfo.addresses, "addresses", []string{},
		"address array of provider, required")
	registerCmd.Flags().StringVar(&ProviderInfo.miner, "miner", "",
		"miner of provider")
	registerCmd.Flags().BoolVarP(&ProviderInfo.onlyEnvelop, "only-envelop", "e", false,
		"only generate envelop body")

	return registerCmd
}

func (f *providerInfo) validateFlags() error {
	if len(f.addresses) == 0 || f.peerID == "" || f.privateKey == "" {
		return fmt.Errorf("all flags are requied, given:\n\taddresses: %v\n\tpeerID%v\n\tprivateKey%v",
			f.addresses, f.peerID, f.privateKey)
	}

	return nil
}
