package mock

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/kenlabs/PandoStore/pkg/config"
	"github.com/kenlabs/PandoStore/pkg/store"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/pkg/policy"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/kenlabs/pando/pkg/registry/discovery"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"time"
)

const topic = "/pando/v0.0.1"

type PandoMock struct {
	Opt      *option.Options
	DS       datastore.Batching
	CS       *badger.DB
	PS       *store.PandoStore
	Host     host.Host
	Core     *legs.Core
	Registry *registry.Registry
	Discover discovery.Discoverer
	outMeta  chan *metadata.MetaRecord
}

func NewPandoMock() (*PandoMock, error) {
	ctx := context.Background()

	_ds := datastore.NewMapDatastore()
	ds := dssync.MutexWrap(_ds)
	cs, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}
	h, err := libp2p.New()
	if err != nil {
		return nil, err
	}
	//bs := blockstore.NewBlockstore(mds)
	ps, err := store.NewStoreFromDatastore(ctx, ds, &config.StoreConfig{
		SnapShotInterval: "5s",
	})
	if err != nil {
		return nil, err
	}

	mockDisco, err := NewMockDiscoverer(exceptID)
	if err != nil {
		return nil, err
	}

	r, err := registry.NewRegistry(ctx, &MockDiscoveryCfg, &MockAclCfg, ds, mockDisco)
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

	opt := option.New(nil)
	_, err = opt.Parse()
	if err != nil {
		return nil, err
	}
	core, err := legs.NewLegsCore(ctx, h, ds, cs, ps, outCh, time.Minute, limiter, r, opt)
	if err != nil {
		return nil, err
	}

	return &PandoMock{
		DS:       ds,
		PS:       ps,
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
