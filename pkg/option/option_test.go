package option

import (
	"github.com/mitchellh/go-homedir"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path/filepath"
	"testing"
)

func TestOptions(t *testing.T) {
	Convey("TestOptions", t, func() {
		opt := New(nil)

		Convey("flag values should be default value before parse function execute", func() {
			So(opt.ConfigFile, ShouldEqual, "config.yaml")
			homeRoot, err := homedir.Expand("~/.pando")
			if err != nil {
				t.Error(err)
			}
			So(opt.PandoRoot, ShouldEqual, homeRoot)
			So(opt.DisableSpeedTest, ShouldEqual, false)
			So(opt.LogLevel, ShouldEqual, "debug")
			So(opt.ServerAddress.HttpAPIListenAddress, ShouldEqual, defaultHttpAPIListenAddress)
			So(opt.ServerAddress.GraphqlListenAddress, ShouldEqual, defaultGraphqlListenAddress)
			So(opt.ServerAddress.P2PAddress, ShouldEqual, defaultP2PAddress)
			So(opt.DataStore.Type, ShouldEqual, defaultDataStoreType)
			So(opt.DataStore.Dir, ShouldEqual, defaultDataStoreDir)
			So(opt.Discovery.Policy.Allow, ShouldEqual, defaultAllow)
			So(opt.Discovery.LotusGateway, ShouldEqual, defaultLotusGateway)
			So(opt.Discovery.Timeout, ShouldEqual, defaultDiscoveryTimeout.String())
			So(opt.Discovery.PollInterval, ShouldEqual, defaultPollInterval.String())
			So(opt.Discovery.RediscoverWait, ShouldEqual, defaultRediscoverWait.String())
			So(opt.AccountLevel.Threshold, ShouldResemble, defaultAccountLevel)
			So(opt.RateLimit.SingleDAGSize, ShouldEqual, defaultSingleDAGSize)
			So(opt.Backup.EstuaryGateway, ShouldEqual, defaultEstGateway)
			So(opt.Backup.ShuttleGateway, ShouldEqual, defaultShuttleGateway)
		})

		Convey("check whether the value of flags are the value set in the specified file", func() {
			opt.PandoRoot = "/tmp/pando"
			opt.ConfigFile = "config.yaml"
			err := os.MkdirAll(opt.PandoRoot, 0755)
			if err != nil {
				t.Error(err)
			}
			file, err := os.OpenFile(filepath.Join(opt.PandoRoot, opt.ConfigFile), os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				t.Error(err)
			}
			_, err = file.WriteString(sampleConfigFileStr())
			if err != nil {
				t.Error(err)
			}
			_, err = opt.Parse()
			if err != nil {
				t.Error()
			}
			So(opt.Identity.PeerID, ShouldEqual, "12D3KooWKw5hu5QcbbFuokt3NrYe7gak5kKHzt8h1FJNqByHQ157")
			So(opt.Identity.PrivateKey, ShouldEqual, "CAESQPWBve9d3ymoWB91XNksRtfEFoiabYd6Qo0qrr+xfbP9lk1SABmWWMwmvv9AiUPELNxExNNCBS18lDkTfMYB5AY=")
			peerID, priavateKey, err := opt.Identity.Decode()
			So(err, ShouldBeNil)
			So(peerID.MatchesPrivateKey(priavateKey), ShouldBeTrue)
			So(opt.ServerAddress.HttpAPIListenAddress, ShouldEqual, "/ip4/0.0.0.0/tcp/8001")
			So(opt.ServerAddress.GraphqlListenAddress, ShouldEqual, "/ip4/0.0.0.0/tcp/8002")
			So(opt.ServerAddress.P2PAddress, ShouldEqual, "/ip4/0.0.0.0/tcp/8003")
			So(opt.RateLimit.Bandwidth, ShouldEqual, 146.81)
			So(opt.Backup.APIKey, ShouldEqual, "EST0933b58d-65f9-470d-bb08-72aed39339f1ARY")
		})

		Convey("return error if privateKey / peerID is invalid", func() {
			// privateKey should be nil, err if it is invalid
			opt.Identity.PeerID = "12D3KooWKw5hu5QcbbFuokt3NrYe7gak5kKHzt8h1FJNqByHQ157"
			opt.Identity.PrivateKey = ""
			_, privateKeyInvalid, err := opt.Identity.Decode()
			So(err, ShouldNotBeNil)
			So(privateKeyInvalid, ShouldBeNil)

			// peerID should be empty,error if it is invalid
			opt.Identity.PeerID = ""
			peerIDInvalid, _, err := opt.Identity.Decode()
			So(err, ShouldNotBeNil)
			So(peerIDInvalid, ShouldBeEmpty)
		})

		err := os.RemoveAll(opt.PandoRoot)
		if err != nil {
			t.Error(err)
		}
	})
}

func sampleConfigFileStr() string {
	return `PandoRoot: /Users/ben/.pando
DisableSpeedTest: false
LogLevel: info
Identity:
  PeerID: 12D3KooWKw5hu5QcbbFuokt3NrYe7gak5kKHzt8h1FJNqByHQ157
  PrivateKey: CAESQPWBve9d3ymoWB91XNksRtfEFoiabYd6Qo0qrr+xfbP9lk1SABmWWMwmvv9AiUPELNxExNNCBS18lDkTfMYB5AY=
ServerAddress:
  HttpAPIListenAddress: /ip4/0.0.0.0/tcp/8001
  GraphqlListenAddress: /ip4/0.0.0.0/tcp/8002
  DisableP2P: false
  P2PAddress: /ip4/0.0.0.0/tcp/8003
DataStore:
  Type: levelds
  Dir: datastore
Discovery:
  Bootstrap: []
  LotusGateway: https://api.chain.love
  Peers: []
  Policy:
    Allow: true
    Except: []
    Trust: true
    TrustExcept: []
  PollInterval: 1s
  RediscoverWait: 1s
  Timeout: 15s
AccountLevel:
  Threshold:
  - 1
  - 10
  - 100
  - 500
RateLimit:
  Bandwidth: 146.81
  SingleDAGSize: 1
Backup:
  EstuaryGateway: https://api.estuary.tech
  ShuttleGateway: https://shuttle-4.estuary.tech
  APIKey: EST0933b58d-65f9-470d-bb08-72aed39339f1ARY`
}
