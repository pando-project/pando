package pando

import (
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const infoPath = "/info"

func infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "show Pando server information",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.Client.R().Get(joinAPIPath(infoPath))
			if err != nil {
				return err
			}

			return api.PrintResponseData(res)
		},
	}

	return cmd
}
