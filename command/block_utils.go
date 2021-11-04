package command

import (
	dag "github.com/ipfs/go-merkledag"

	bsrv "github.com/ipfs/go-blockservice"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipld "github.com/ipfs/go-ipld-format"
)

// Mock returns a new thread-safe DAGService.
func BlockStoreFromDataStore(ds *dssync.MutexDatastore) ipld.DAGService {
	return dag.NewDAGService(Bserv(ds))
}

// Bserv returns a new, thread-safe BlockService.
func Bserv(datastore *dssync.MutexDatastore) bsrv.BlockService {
	bstore := blockstore.NewBlockstore(datastore)
	return bsrv.New(bstore, offline.Exchange(bstore))
}
