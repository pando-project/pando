module github.com/kenlabs/pando

go 1.16

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/agiledragon/gomonkey/v2 v2.4.0
	github.com/briandowns/spinner v1.11.1
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/fatih/color v1.9.0 // indirect
	github.com/filecoin-project/go-address v0.0.6 // indirect
	github.com/filecoin-project/go-bitfield v0.2.4 // indirect
	github.com/filecoin-project/go-cbor-util v0.0.1 // indirect
	github.com/filecoin-project/go-data-transfer v1.14.0
	github.com/filecoin-project/go-indexer-core v0.2.7
	github.com/filecoin-project/go-legs v0.3.7
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/go-statemachine v1.0.1 // indirect
	github.com/filecoin-project/specs-actors v0.9.14 // indirect
	github.com/filecoin-project/specs-actors/v2 v2.3.5 // indirect
	github.com/filecoin-project/specs-actors/v3 v3.1.1 // indirect
	github.com/filecoin-project/specs-actors/v5 v5.0.4
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-openapi/runtime v0.21.0
	github.com/go-resty/resty/v2 v2.7.0
	github.com/graphql-go/graphql v0.8.0
	github.com/gwatts/gin-adapter v0.0.0-20170508204228-c44433c485ad
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-blockservice v0.2.1
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-datastore v0.5.1
	github.com/ipfs/go-ds-leveldb v0.5.0
	github.com/ipfs/go-graphsync v0.12.0
	github.com/ipfs/go-ipfs-blockstore v1.1.2
	github.com/ipfs/go-ipfs-exchange-offline v0.1.1
	github.com/ipfs/go-ipfs-files v0.0.9 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.6
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.5.0
	github.com/ipfs/go-merkledag v0.5.1
	github.com/ipld/go-car/v2 v2.1.1
	github.com/ipld/go-ipld-prime v0.16.0
	//github.com/ipld/go-ipld-prime v0.14.4
	github.com/libp2p/go-libp2p v0.18.0-rc1
	github.com/libp2p/go-libp2p-core v0.14.0
	github.com/libp2p/go-libp2p-record v0.1.3 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/montanaflynn/stats v0.6.6
	github.com/multiformats/go-multiaddr v0.5.0
	github.com/multiformats/go-multicodec v0.4.0
	github.com/multiformats/go-multihash v0.1.0
	github.com/prometheus/client_golang v1.11.0
	github.com/shopspring/decimal v1.3.1
	github.com/showwin/speedtest-go v1.1.4
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/warpfork/go-testmark v0.9.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20220224212727-7a699437a831
	go.elastic.co/apm/module/apmhttp v1.15.0
	go.opencensus.io v0.23.0
	go.uber.org/goleak v1.1.11 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.1.8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	//github.com/ipld/go-ipld-prime v0.14.4 => github.com/ipld/go-ipld-prime v0.14.3
	github.com/libp2p/go-libp2p v0.18.0-rc1 => github.com/libp2p/go-libp2p v0.17.0
	github.com/libp2p/go-libp2p-core v0.14.0 => github.com/libp2p/go-libp2p-core v0.13.0
	github.com/showwin/speedtest-go v1.1.4 => github.com/kenlabs/speedtest-go v1.1.5
)
