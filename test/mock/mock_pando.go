package mock

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"pando/pkg/legs"
	"pando/pkg/metadata"
	"pando/pkg/option"
	"pando/pkg/policy"
	"pando/pkg/registry"
	"pando/pkg/registry/discovery"
	"time"
)

const topic = "/pando/v0.0.1"

type PandoMock struct {
	Opt      *option.Options
	DS       datastore.Batching
	CS       *badger.DB
	BS       blockstore.Blockstore
	Host     host.Host
	Core     *legs.Core
	Registry *registry.Registry
	Discover discovery.Discoverer
	outMeta  chan *metadata.MetaRecord
}

func NewPandoMock() (*PandoMock, error) {
	ctx := context.Background()

	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	cs, err := badger.Open(badger.DefaultOptions("/tmp"))
	if err != nil {
		return nil, err
	}
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}
	bs := blockstore.NewBlockstore(mds)

	mockDisco, err := NewMockDiscoverer(exceptID)
	if err != nil {
		return nil, err
	}

	r, err := registry.NewRegistry(ctx, &MockDiscoveryCfg, &MockAclCfg, mds, mockDisco)
	if err != nil {
		return nil, err
	}

	limiter, err := policy.NewLimiter(policy.LimiterConfig{
		TotalRate:     BaseTokenRate,
		TotalBurst:    int(BaseTokenRate),
		Registry:      r,
		BaseTokenRate: BaseTokenRate,
	})
	if err != nil {
		return nil, err
	}

	outCh := make(chan *metadata.MetaRecord)
	core, err := legs.NewLegsCore(ctx, h, mds, cs, bs, outCh, time.Minute, limiter, r)
	if err != nil {
		return nil, err
	}
	//r.SetCore(core)
	opt := option.New(nil)
	_, err = opt.Parse()
	if err != nil {
		return nil, err
	}

	return &PandoMock{
		DS:       mds,
		BS:       bs,
		CS:       cs,
		Host:     h,
		Core:     core,
		Registry: r,
		Discover: mockDisco,
		outMeta:  outCh,
		Opt:      opt,
	}, nil
}

func (pando *PandoMock) GetMetaRecordCh() (chan *metadata.MetaRecord, error) {
	if pando.outMeta != nil {
		return pando.outMeta, nil
	}
	return nil, fmt.Errorf("nil channel")
}

func GetTopic() string {
	return topic
}
