package command

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var PandoAPI string
var PandoClient = resty.New().SetBaseURL(PandoAPI)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:        "pando",
		Short:      "Pando client cli",
		SuggestFor: []string{"pando"},
	}

	rootCmd.PersistentFlags().StringVarP(&PandoAPI, "pando-api", "a", "",
		"set pando api url")

	childCommands := []*cobra.Command{
		RegisterCmd(),
	}
	rootCmd.AddCommand(childCommands...)

	return rootCmd
}
