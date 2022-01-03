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
	PollInterval   string          `yaml:"PollInterval"`
	RediscoverWait string          `yaml:"RediscoverWait"`
	Timeout        string          `yaml:"Timeout"`
}

func (d *Discovery) PollIntervalInDurationFormat() Duration {
	return unmarshalDurationString(d.PollInterval)
}

func (d *Discovery) RediscoverWaitInDurationFormat() Duration {
	return unmarshalDurationString(d.RediscoverWait)
}

func (d *Discovery) TimeoutInDurationFormat() Duration {
	return unmarshalDurationString(d.Timeout)
}

func unmarshalDurationString(durationStr string) Duration {
	d := Duration(0)
	_ = d.UnmarshalText([]byte(durationStr))
	return d
}
