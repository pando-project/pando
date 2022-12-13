package core

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/ipfs/go-datastore/sync"
	"github.com/ipld/go-ipld-prime"
	"github.com/kenlabs/pando-store/pkg/store"
	"github.com/pando-project/pando/pkg/legs"
	"github.com/pando-project/pando/pkg/lotus"
	"github.com/pando-project/pando/pkg/metadata"
	"github.com/pando-project/pando/pkg/registry"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel/trace"
)

type Core struct {
	MetaManager *metadata.MetaManager
	//StateTree     *statetree.StateTree
	LotusDiscover *lotus.Discoverer
	Registry      *registry.Registry
	LegsCore      *legs.Core
	StoreInstance *StoreInstance
	LinkSystem    *ipld.LinkSystem

	TraceCtx  *context.Context
	TraceSpan *trace.Span
}

type StoreInstance struct {
	//DataStore      *leveldb.Datastore
	MutexDataStore *sync.MutexDatastore
	//BlockStore     blockstore.Blockstore
	PandoStore    *store.PandoStore
	CacheStore    *badger.DB
	MetadataCache *mongo.Client
}
