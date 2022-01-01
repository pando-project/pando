package provider

import (
	"github.com/spf13/cobra"
	"pando/cmd/client/command/api"
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
