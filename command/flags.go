package command

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
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

var pandoHostFlag = altsrc.NewStringFlag(&cli.StringFlag{
	Name:     "pando",
	Usage:    "Host or host:port of pando to use",
	EnvVars:  []string{"PANDO"},
	Required: false,
	Value:    "localhost",
})

var registerFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "config",
		Usage:    "Config file containing provider's peer ID and private key",
		Required: true,
	},
	pandoHostFlag,
	&cli.StringSliceFlag{
		Name:     "provider-addr",
		Usage:    "Provider address as multiaddr string, example: \"/ip4/127.0.0.1/tcp/3102\"",
		Aliases:  []string{"pa"},
		Required: true,
	},
}
