package command

import (
	"github.com/spf13/cobra"
	"pando/pkg/option"
	"pando/pkg/system"
)

var Opt *option.Options

var ExampleUsage = `
# Init Pando configs.
pando-server init

# Start pando server.
pando-server start

# Start pando server in daemon mode.
pando-server start -d
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
