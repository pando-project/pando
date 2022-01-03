package core

import (
	"github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-ds-leveldb"
	"github.com/ipfs/go-ipfs-blockstore"
	"pando/pkg/legs"
	"pando/pkg/lotus"
	"pando/pkg/metadata"
	"pando/pkg/registry"
	"pando/pkg/statetree"
)

type Core struct {
	MetaManager   *metadata.MetaManager
	StateTree     *statetree.StateTree
	LotusDiscover *lotus.Discoverer
	Registry      *registry.Registry
	LegsCore      *legs.Core
	StoreInstance *StoreInstance
}

type StoreInstance struct {
	DataStore      *leveldb.Datastore
	MutexDataStore *sync.MutexDatastore
	BlockStore     blockstore.Blockstore
}