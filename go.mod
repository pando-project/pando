module pando

go 1.16

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/agiledragon/gomonkey/v2 v2.3.1
	github.com/briandowns/spinner v1.11.1
	github.com/filecoin-project/go-address v0.0.6
	github.com/filecoin-project/go-cbor-util v0.0.1 // indirect
	github.com/filecoin-project/go-commp-utils v0.1.3 // indirect
	github.com/filecoin-project/go-data-transfer v1.11.4
	github.com/filecoin-project/go-fil-markets v1.13.3 // indirect
	github.com/filecoin-project/go-indexer-core v0.2.6
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-legs v0.0.0-20211116112108-61960ef1f8ef
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/lotus v1.13.0
	github.com/filecoin-project/specs-actors/v5 v5.0.4
	github.com/filecoin-project/specs-actors/v6 v6.0.1 // indirect
	github.com/gammazero/keymutex v0.0.2
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-openapi/runtime v0.21.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/graphql-go/graphql v0.8.0
	github.com/gwatts/gin-adapter v0.0.0-20170508204228-c44433c485ad
	github.com/ipfs/go-bitswap v0.4.0 // indirect
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-blockservice v0.1.7
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-datastore v0.4.6
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-graphsync v0.10.5
	github.com/ipfs/go-ipfs-blockstore v1.0.5-0.20210802214209-c56038684c45
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipfs-files v0.0.9 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/ipfs/go-merkledag v0.4.1
	github.com/ipld/go-car v0.3.2
	github.com/ipld/go-ipld-prime v0.12.4-0.20211026094848-168715526f2d
	github.com/justinas/nosurf v1.1.1 // indirect
	github.com/libp2p/go-libp2p v0.15.1
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/montanaflynn/stats v0.6.6
	github.com/multiformats/go-multiaddr v0.4.1
	github.com/multiformats/go-multicodec v0.3.1-0.20210902112759-1539a079fd61
	github.com/prometheus/client_golang v1.11.0
	github.com/shopspring/decimal v1.3.1
	github.com/showwin/speedtest-go v1.1.4
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.7.0
	github.com/turtlemonvh/gin-wraphh v0.0.0-20160304035037-ea8e4927b3a6
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	go.opencensus.io v0.23.0
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/showwin/speedtest-go v1.1.4 => github.com/kenlabs/speedtest-go v1.1.5
