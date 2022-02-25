package metadata

import (
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/metadata"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewMetadataCmd() *cobra.Command {
	metadataCmd := &cobra.Command{
		Use:   "metadata",
		Short: "show metadata information",
	}

	childCommands := []*cobra.Command{
		listCmd(),
		snapshotCmd(),
	}
	metadataCmd.AddCommand(childCommands...)

	return metadataCmd
}
