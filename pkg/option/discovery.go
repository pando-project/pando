package option

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	defaultLotusGateway     = "https://api.chain.love"
	defaultPollInterval     = Duration(24 * time.Hour)
	defaultRediscoverWait   = Duration(5 * time.Minute)
	defaultDiscoveryTimeout = Duration(2 * time.Minute)
)

type Discovery struct {
	Bootstrap      []string        `yaml:"Bootstrap"`
	LotusGateway   string          `yaml:"LotusGateway"`
	Peers          []peer.AddrInfo `yaml:"Peers"`
	Policy         Policy          `yaml:"Policy"`
	PollInterval   Duration        `yaml:"PollInterval"`
	RediscoverWait Duration        `yaml:"RediscoverWait"`
	Timeout        Duration        `yaml:"Timeout"`
}
