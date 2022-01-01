package pando

import (
	"github.com/spf13/cobra"
	"pando/cmd/client/command/api"
)

const healthPath = "/health"

func healthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check if Pando server is alive",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.Client.R().Options(joinAPIPath(healthPath))
			if err != nil {
				return err
			}

			return api.PrintResponseData(res)
		},
	}

	return cmd
}
