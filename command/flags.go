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
	&cli.BoolFlag{
		Name:     "speedtest",
		Usage:    "switch of speedtest",
		Value:    true,
		Required: false,
	},
}

var pandoHostFlag = altsrc.NewStringFlag(&cli.StringFlag{
	Name:     "pando",
	Usage:    "Pando service address, which accepts hostname or hostname:port",
	EnvVars:  []string{"PANDO"},
	Required: false,
	Value:    "localhost",
})

var registerFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "config",
		Usage:    "Config file containing provider's account ID and private key",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "privkey",
		Usage:    "Provider private key, used for signing the register request",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "peerid",
		Usage:    "Provider peer ID",
		Required: false,
	},
	pandoHostFlag,
	&cli.StringSliceFlag{
		Name:     "provider-addr",
		Usage:    "Provider multiaddr, example: \"/ip4/127.0.0.1/tcp/3102\"",
		Aliases:  []string{"pa"},
		Required: true,
	},
	&cli.StringFlag{
		Name:     "miner",
		Usage:    "Provider miner account (optional)",
		Required: false,
	},
}

var ingestFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "provider",
		Usage:    "Provider to interact with",
		Aliases:  []string{"prov"},
		Required: true,
	},
	pandoHostFlag,
}
