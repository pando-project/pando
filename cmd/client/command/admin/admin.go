package admin

import (
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/admin"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "admin the Pando",
	}

	childCommands := []*cobra.Command{
		backupCmd(),
	}
	adminCmd.AddCommand(childCommands...)

	return adminCmd
}
