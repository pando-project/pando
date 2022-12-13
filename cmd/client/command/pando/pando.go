package pando

import (
	"github.com/pando-project/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/pando"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewPandoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pando",
		Short: "show information of Pando server",
	}

	childCommands := []*cobra.Command{
		infoCmd(),
		healthCmd(),
	}
	cmd.AddCommand(childCommands...)

	return cmd
}
