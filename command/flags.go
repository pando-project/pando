package command

import (
	"github.com/urfave/cli/v2"
)

var initFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "listen-graphsync",
		Usage:    "GraphSync HTTP API listen address",
		EnvVars:  []string{"PANDO_LISTEN_GRAPHSYNC"},
		Required: false,
	},
	&cli.StringFlag{
		Name:     "listen-graphql",
		Usage:    "GraphQl HTTP API listen address",
		EnvVars:  []string{"PANDO_LISTEN_GRAPHQL"},
		Required: false,
	},
}

var snapShotFlags = []cli.Flag{
	&cli.StringFlag{
		Name: "ls",
		//Usage:    "GraphSync HTTP API listen address",
		//EnvVars:  []string{"PANDO_LISTEN_GRAPHSYNC"},
		Required: false,
	},
	&cli.StringFlag{
		Name: "",
		//Usage:    "GraphQl HTTP API listen address",
		//EnvVars:  []string{"PANDO_LISTEN_GRAPHQL"},
		Required: false,
	},
}
