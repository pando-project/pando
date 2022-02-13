package provider

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"

	"pando/cmd/client/command/api"
	"pando/pkg/register"
)

const registerPath = "/register"

type providerInfo struct {
	peerID      string
	privateKey  string
	addresses   []string
	miner       string
	onlyEnvelop bool
}

var registerInfo = &providerInfo{}

func registerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "register a provider to pando server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := registerInfo.validateFlags(); err != nil {
				return err
			}

			peerID, err := peer.Decode(registerInfo.peerID)
			if err != nil {
				return err
			}

			privateKeyEncoded, err := base64.StdEncoding.DecodeString(registerInfo.privateKey)
			if err != nil {
				return err
			}
			privateKey, err := crypto.UnmarshalPrivateKey(privateKeyEncoded)
			if err != nil {
				return err
			}

			data, err := register.MakeRegisterRequest(peerID, privateKey, registerInfo.addresses, registerInfo.miner)
			if err != nil {
				return err
			}

			if registerInfo.onlyEnvelop {
				envelopFile, err := os.OpenFile("./envelop.data", os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return err
				}
				_, err = envelopFile.Write(data)
				if err != nil {
					return err
				}
				fmt.Println("envelop data saved at ./envelop.data")
				return nil
			}

			res, err := api.Client.R().
				SetBody(data).
				SetHeader("Content-Type", "application/octet-stream").
				Post(joinAPIPath(registerPath))
			if err != nil {
				return err
			}
			return api.PrintResponseData(res)
		},
	}

	registerInfo.setFlags(cmd)

	return cmd
}

func (f *providerInfo) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.peerID, "peer-id", "",
		"peerID of provider, required")
	cmd.Flags().StringVar(&f.privateKey, "private-key", "",
		"private key of provider, required")
	cmd.Flags().StringSliceVar(&f.addresses, "addresses", []string{},
		"address array of provider")
	cmd.Flags().StringVar(&f.miner, "miner", "",
		"miner of provider")
	cmd.Flags().BoolVarP(&f.onlyEnvelop, "only-envelop", "e", false,
		"only generate envelop body")
}

func (f *providerInfo) validateFlags() error {
	if f.peerID == "" || f.privateKey == "" {
		return fmt.Errorf("peerID and privateKey are requied, given:\n\taddresses: %v\n\tpeerID%v\n\tprivateKey%v",
			f.addresses, f.peerID, f.privateKey)
	}

	return nil
}
