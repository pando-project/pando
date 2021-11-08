package command

import "github.com/urfave/cli/v2"

var SnapShotCmd = &cli.Command{
	Name:   "snapshot",
	Usage:  "get info about snapshot",
	Flags:  nil,
	Action: daemonCommand,
	Subcommands: []*cli.Command{
		listSnapShots,
	},
}

var listSnapShots = &cli.Command{
	Name:   "cidlist",
	Usage:  "Import indexer data from cidList",
	Flags:  nil,
	Action: listSnapShotsCommand,
}

func listSnapShotsCommand(cctx *cli.Context) error {
	return nil
}
