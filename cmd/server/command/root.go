package command

import (
	"github.com/pando-project/pando/pkg/option"
	"github.com/pando-project/pando/pkg/system"
	"github.com/spf13/cobra"
)

var Opt *option.DaemonOptions

var ExampleUsage = `
# Init Pando configs(default path is ~/.pando/config.yaml).
pando-server init

# StartHttpServer pando server.
pando-server daemon
`

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:        "pando-server",
		Short:      "Pando server cli",
		Example:    ExampleUsage,
		SuggestFor: []string{"pando-server"},
	}

	Opt = option.New(rootCmd)
	msg, err := Opt.Parse()
	if err != nil {
		system.Exit(1, err.Error())
	}
	if msg != "" {
		system.Exit(0, msg)
	}

	childCommands := []*cobra.Command{
		InitCmd(),
		DaemonCmd(),
	}
	rootCmd.AddCommand(childCommands...)

	return rootCmd
}
