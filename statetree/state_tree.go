package statetree

import (
	"Pando/statetree/hamt"
	"Pando/statetree/types"
	"context"
	"fmt"
	"github.com/filecoin-project/specs-actors/v5/actors/builtin"
	"github.com/filecoin-project/specs-actors/v5/actors/util/adt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
)

const RootKey = "PandoStateTree"

var log = logging.Logger("state-tree")

type StateTree struct {
	ds    datastore.Batching
	Store cbor.IpldStore
	root  hamt.Map
	//state        map[peer.ID]*types.ProviderState
	recvUpdateCh <-chan map[peer.ID]*types.ProviderState
}

type SnapShot struct {
	State      map[peer.ID]types.ProviderState
	Height     uint64
	CreateTime uint64
	Previous   cid.Cid
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore, updateCh <-chan map[peer.ID]*types.ProviderState) (*StateTree, error) {
	cs := cbor.NewCborStore(bs)
	store := adt.WrapStore(context.TODO(), cs)
	st := &StateTree{
		ds:           ds,
		Store:        store,
		recvUpdateCh: updateCh,
	}
	// get the latest root(cid) from ds
	root, err := ds.Get(datastore.NewKey(RootKey))
	if err == nil && root != nil {
		_, rootcid, err := cid.CidFromBytes(root)
		if err != nil {
			return nil, fmt.Errorf("failed to load the State root from datastore")
		}
		log.Debugf("find root cid %s, loading...", rootcid.String())
		//store := adt.WrapStore(context.TODO(), cst)

		m, err := adt.AsMap(store, rootcid, builtin.DefaultHamtBitwidth)
		if err != nil {
			return nil, fmt.Errorf("failed to create hamt root by cid: %s\r\n%s", rootcid.String(), err.Error())
		}
		st.root = m
	} else {
		emptyRoot, err := adt.MakeEmptyMap(store, builtin.DefaultHamtBitwidth)
		if err != nil {
			return nil, err
		}
		st.root = emptyRoot
	}
	// if MetadataManager send metadata update, update the State, save in hamt and change the root cid
	go st.Update()

	return st, nil
}

func (st *StateTree) Update() {
	for {
		select {
		case update, ok := <-st.recvUpdateCh:
			if !ok {
				log.Error("metadata manager close the update channel.")
				return
			}
			err := st.UpdateRoot(update)
			if err != nil {
				log.Errorf("while updating the state tree, some errors happened!\r\n%s", err.Error())
			}
		}
	}
}

func (st *StateTree) UpdateRoot(update map[peer.ID]*types.ProviderState) error {
	log.Debug("start updating the state tree")
	for p, cidlist := range update {
		state := new(types.ProviderState)
		found, err := st.root.Get(hamt.ProviderKey{ID: p}, state)
		if err != nil {
			return fmt.Errorf("failed to get provider state from hamt, %s", err.Error())
		}
		if !found {
			err = st.root.Put(hamt.ProviderKey{ID: p}, &types.ProviderState{Cidlist: cidlist.Cidlist})
			if err != nil {
				return fmt.Errorf("failed to put provider state into hamt, %s", err.Error())
			}
		} else {
			state.Cidlist = append(state.Cidlist, cidlist.Cidlist...)
			err = st.root.Put(hamt.ProviderKey{ID: p}, state)
			if err != nil {
				return fmt.Errorf("failed to put provider state into hamt, %s", err.Error())
			}
		}
		newRootCid, err := st.root.Root()
		if err != nil {
			return fmt.Errorf("failed to put new hamt root into datastore")
		}
		// put the latest root(cid) into ds
		log.Debugf("saving the new root, cid: %s", newRootCid.String())
		err = st.ds.Put(datastore.NewKey(RootKey), newRootCid.Bytes())
		if err != nil {
			return fmt.Errorf("failed to save the newest root cid")
		}

	}
	return nil
}
