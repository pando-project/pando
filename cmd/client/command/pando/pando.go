package pando

import (
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/pando"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewPandoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pando",
		Short: "let Pando subscribe a topic to start synchronization, show information of Pando server",
	}

	childCommands := []*cobra.Command{
		//subscribeCmd(),
		infoCmd(),
		healthCmd(),
	}
	cmd.AddCommand(childCommands...)

	return cmd
}
