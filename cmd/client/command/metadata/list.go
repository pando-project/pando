package metadata

import (
	"github.com/pando-project/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const listPath = "/list"

func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "return a list of metadata snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.Client.R().Get(joinAPIPath(listPath))
			if err != nil {
				return err
			}
			return api.PrintResponseData(res)
		},
	}

	return cmd
}
