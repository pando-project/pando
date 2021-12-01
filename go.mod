module Pando

go 1.16

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0
	github.com/agiledragon/gomonkey/v2 v2.3.1
	github.com/briandowns/spinner v1.11.1
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-indexer-core v0.2.6
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-legs v0.0.0-20211116112108-61960ef1f8ef
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/lotus v1.13.0
	github.com/filecoin-project/specs-actors/v5 v5.0.4
	github.com/gammazero/keymutex v0.0.2
	github.com/gorilla/mux v1.8.0
	github.com/graphql-go/graphql v0.8.0
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-blockservice v0.1.5
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-datastore v0.4.6
	github.com/ipfs/go-ds-leveldb v0.4.2
	github.com/ipfs/go-graphsync v0.10.1
	github.com/ipfs/go-ipfs-blockstore v1.0.4
	github.com/ipfs/go-ipfs-exchange-offline v0.0.1
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-ipld-format v0.2.0
	github.com/ipfs/go-log/v2 v2.3.0
	github.com/ipfs/go-merkledag v0.3.2
	github.com/ipld/go-car v0.3.2
	github.com/ipld/go-ipld-prime v0.12.4-0.20211026094848-168715526f2d
	github.com/libp2p/go-libp2p v0.15.1
	github.com/libp2p/go-libp2p-core v0.9.0
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/montanaflynn/stats v0.6.6
	github.com/multiformats/go-multiaddr v0.4.0
	github.com/multiformats/go-multicodec v0.3.1-0.20210902112759-1539a079fd61
	github.com/prometheus/client_golang v1.11.0
	github.com/shopspring/decimal v1.3.1
	github.com/showwin/speedtest-go v1.1.4
	github.com/smartystreets/goconvey v1.6.4
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20210713220151-be142a5ae1a8
	go.opencensus.io v0.23.0
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/showwin/speedtest-go v1.1.4 => github.com/kenlabs/speedtest-go v1.1.5
