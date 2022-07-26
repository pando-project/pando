package core

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/ipfs/go-datastore/sync"
	"github.com/ipld/go-ipld-prime"
	"github.com/kenlabs/pando-store/pkg/store"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/lotus"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/registry"
	"go.mongodb.org/mongo-driver/mongo"
)

type Core struct {
	MetaManager *metadata.MetaManager
	//StateTree     *statetree.StateTree
	LotusDiscover *lotus.Discoverer
	Registry      *registry.Registry
	LegsCore      *legs.Core
	StoreInstance *StoreInstance
	LinkSystem    *ipld.LinkSystem
}

type StoreInstance struct {
	//DataStore      *leveldb.Datastore
	MutexDataStore *sync.MutexDatastore
	//BlockStore     blockstore.Blockstore
	PandoStore    *store.PandoStore
	CacheStore    *badger.DB
	MetadataCache *mongo.Client
}
