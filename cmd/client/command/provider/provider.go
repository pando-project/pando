package provider

import (
	"github.com/pando-project/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/provider"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "provider management commands",
	}

	childCommands := []*cobra.Command{
		registerCmd(),
	}
	cmd.AddCommand(childCommands...)

	return cmd
}
