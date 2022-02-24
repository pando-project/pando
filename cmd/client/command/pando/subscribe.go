package pando

import (
	"fmt"
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
)

const subscribePath = "/subscribe"

type subscribeInfo struct {
	peerID string
}

var providerSubInfo = &subscribeInfo{}

func subscribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "let Pando subscribe a topic to start synchronization with provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := providerSubInfo.validateFlags()
			if err != nil {
				return err
			}

			res, err := api.Client.R().
				SetQueryParam("provider", providerSubInfo.peerID).
				Get(joinAPIPath(subscribePath))
			if err != nil {
				return err
			}
			return api.PrintResponseData(res)
		},
	}
	providerSubInfo.setFlags(cmd)

	return cmd
}

func (p *subscribeInfo) setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.peerID, "provider-peerid", "",
		"set provider peer id to make Pando subscribe and start metadata synchronization")
}

func (p *subscribeInfo) validateFlags() error {
	if p.peerID == "" {
		return fmt.Errorf("peer id of provider is empty")
	}
	_, err := peer.Decode(p.peerID)
	if err != nil {
		return fmt.Errorf("invalide peer id: %v", err)
	}

	return nil
}
