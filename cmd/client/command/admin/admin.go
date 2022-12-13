package admin

import (
	"github.com/pando-project/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const groupPath = "/"

var joinAPIPath = api.JoinPathFuncFactory(groupPath)

func NewAdminCmd() *cobra.Command {
	adminCmd := &cobra.Command{
		Use:   "admin",
		Short: "Pando administration",
	}

	childCommands := []*cobra.Command{
		backupCmd(),
	}
	adminCmd.AddCommand(childCommands...)

	return adminCmd
}
