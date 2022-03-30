package option

import (
	"time"
)

const (
	defaultLotusGateway     = "https://api.chain.love"
	defaultPollInterval     = Duration(24 * time.Hour)
	defaultPollRetryAfter   = Duration(5 * time.Hour)
	defaultPollStopAfter    = Duration(7 * 24 * time.Hour)
	defaultRediscoverWait   = Duration(5 * time.Minute)
	defaultDiscoveryTimeout = Duration(2 * time.Minute)
)

type Discovery struct {
	//Bootstrap      []string        `yaml:"Bootstrap"`
	LotusGateway string `yaml:"LotusGateway"`
	//Peers          []peer.AddrInfo `yaml:"Peers"`
	// PollInterval is the amount of time to wait without getting any updates
	// for a provider, before sending a request for the latest advertisement.
	// Values are a number ending in "s", "m", "h" for seconds. minutes, hours.
	PollInterval string `yaml:"PollInterval"`
	// PollRetryAfter is the amount of time from one poll attempt, without a
	// response, to the next poll attempt, and is also the time between checks
	// for providers to poll.  This value must be smaller than PollStopAfter
	// for there to be more than one poll attempt for a provider.
	PollRetryAfter string `yaml:"PollRetryAfter"`
	// PollStopAfter is the amount of time, from the start of polling, to
	// continuing polling for the latest advertisment without getting a
	// responce.
	PollStopAfter  string `yaml:"PollStopAfter"`
	Policy         Policy `yaml:"Policy"`
	RediscoverWait string `yaml:"RediscoverWait"`
	Timeout        string `yaml:"Timeout"`
}

func (d *Discovery) PollIntervalInDurationFormat() Duration {
	return unmarshalDurationString(d.PollInterval)
}

func (d *Discovery) PollRetryAfterInDurationFormat() Duration {
	return unmarshalDurationString(d.PollRetryAfter)
}

func (d *Discovery) PollStopAfterInDurationFormat() Duration {
	return unmarshalDurationString(d.PollStopAfter)
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
