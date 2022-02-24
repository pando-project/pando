package pando

import (
	"fmt"
	"github.com/kenlabs/pando/cmd/client/command/api"
	"github.com/spf13/cobra"
)

const healthPath = "/health"

func healthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check if Pando server is alive",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := api.Client.R().Options(joinAPIPath(healthPath))
			if err != nil {
				return err
			}

			if res.StatusCode() >= 200 && res.StatusCode() < 300 {
				fmt.Println("I'm good.")
			} else {
				fmt.Println("Not healthy.")
			}

			return nil
		},
	}

	return cmd
}
