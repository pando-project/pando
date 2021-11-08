package statetree

import (
	"Pando/statetree/hamt"
	"Pando/statetree/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/specs-actors/v5/actors/builtin"
	"github.com/filecoin-project/specs-actors/v5/actors/util/adt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"time"
)

const (
	// RootKey saves the hamt root cid in ds
	RootKey = "PandoStateTree"
	// SnapShotList saves the snapshots' cids in ds (bytes of []cid.Cid , not save cid of cidlist)
	SnapShotList = "PandoSnapShotList"
)

var log = logging.Logger("state-tree")

type SnapShotCidList []cid.Cid

type StateTree struct {
	ds       datastore.Batching
	Store    cbor.IpldStore
	root     hamt.Map
	snapShot cid.Cid
	exinfo   *types.ExtraInfo
	//state        map[peer.ID]*types.ProviderState
	recvUpdateCh <-chan map[peer.ID]*types.ProviderState
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore, updateCh <-chan map[peer.ID]*types.ProviderState, exinfo *types.ExtraInfo) (*StateTree, error) {
	cs := cbor.NewCborStore(bs)
	store := adt.WrapStore(context.TODO(), cs)
	st := &StateTree{
		exinfo:       exinfo,
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

	snapShotCidList, err := st.GetSnapShotCidList()
	if err != nil {
		return nil, err
	}
	if snapShotCidList != nil {
		ss := new(types.SnapShot)
		newestSsCid := snapShotCidList[len(snapShotCidList)-1]
		err = store.Get(ctx, newestSsCid, ss)
		if err != nil {
			return nil, fmt.Errorf("failed to load the newest snapshot: %s", newestSsCid.String())
		}
		st.snapShot = newestSsCid
	} else {
		st.snapShot = cid.Undef
	}

	//snapShotList, err := ds.Get(datastore.NewKey(SnapShotList))
	//if err == nil && snapShotList != nil {
	//	ssCidList := new(SnapShotCidList)
	//	err := json.Unmarshal(snapShotList, ssCidList)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to load the snapshot cid list from datastore")
	//	}
	//
	//	ss := new(types.SnapShot)
	//	newestSsCid := (*ssCidList)[len(*ssCidList)-1]
	//	err = store.Get(ctx, newestSsCid, ss)
	//	if err != nil {
	//		return nil, fmt.Errorf("failed to load the newest snapshot: %s", newestSsCid.String())
	//	}
	//	st.snapShot = newestSsCid
	//} else {
	//	st.snapShot = cid.Undef
	//}

	// if MetadataManager send metadata update, update the State, save in hamt and change the root cid
	go st.Update(ctx)

	return st, nil
}

func (st *StateTree) Update(ctx context.Context) {
	for {
		select {
		case update, ok := <-st.recvUpdateCh:
			if !ok {
				log.Error("metadata manager close the update channel.")
				return
			}
			rootCid, err := st.UpdateRoot(ctx, update)
			if err != nil {
				log.Errorf("while updating the state tree, some errors happened!\r\n%s", err.Error())
			}
			err = st.CreateSnapShot(ctx, rootCid, update)
			if err != nil {
				log.Errorf("while creating and saving snapshot, some errors happened!\r\n%s", err.Error())
			}
		}
	}
}

func (st *StateTree) UpdateRoot(ctx context.Context, update map[peer.ID]*types.ProviderState) (cid.Cid, error) {
	log.Debug("start updating the state tree")
	for p, cidlist := range update {
		state := new(types.ProviderState)
		found, err := st.root.Get(hamt.ProviderKey{ID: p}, state)
		if err != nil {
			return cid.Undef, fmt.Errorf("failed to get provider state from hamt, %s", err.Error())
		}
		if !found {
			err = st.root.Put(hamt.ProviderKey{ID: p}, &types.ProviderState{Cidlist: cidlist.Cidlist})
			if err != nil {
				return cid.Undef, fmt.Errorf("failed to put provider state into hamt, %s", err.Error())
			}
		} else {
			state.Cidlist = append(state.Cidlist, cidlist.Cidlist...)
			err = st.root.Put(hamt.ProviderKey{ID: p}, state)
			if err != nil {
				return cid.Undef, fmt.Errorf("failed to put provider state into hamt, %s", err.Error())
			}
		}
	}
	newRootCid, err := st.root.Root()
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to put new hamt root into datastore")
	}
	// put the latest root(cid) into ds
	log.Debugf("saving the new root, cid: %s", newRootCid.String())
	err = st.ds.Delete(datastore.NewKey(RootKey))
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to clean the old root cid")
	}
	err = st.ds.Put(datastore.NewKey(RootKey), newRootCid.Bytes())
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to save the newest root cid")
	}

	return newRootCid, nil
}

func (st *StateTree) CreateSnapShot(ctx context.Context, newRoot cid.Cid, update map[peer.ID]*types.ProviderState) error {
	var height uint64
	var previousSs cid.Cid
	if st.snapShot == cid.Undef {
		// todo
		height = uint64(1)
		previousSs = newRoot
	} else {
		oldSs := new(types.SnapShot)
		err := st.Store.Get(ctx, st.snapShot, oldSs)
		if err != nil {
			return fmt.Errorf("failed to load the old snapshot root from datastore")
		}
		height = oldSs.Height + 1
		previousSs = st.snapShot
	}

	_update := make(map[string]*types.ProviderState)
	for p, cidlist := range update {
		_update[p.String()] = cidlist
	}

	newSs := &types.SnapShot{
		Update:     _update,
		Height:     height,
		CreateTime: uint64(time.Now().UnixNano()),
		Root:       newRoot,
		PreviousSs: previousSs,
		ExtraInfo:  st.exinfo,
	}

	ssCid, err := st.Store.Put(ctx, newSs)
	if err != nil {
		return fmt.Errorf("falied to save the snapshot. %s", err.Error())
	}
	st.snapShot = ssCid
	err = st.UpdateSnapShotCidList(ssCid)
	if err != nil {
		return fmt.Errorf("error happened while updating the cidlist of snapshot.%s", err.Error())
	}

	return nil
}

func (st *StateTree) UpdateSnapShotCidList(newSsCid cid.Cid) error {
	snapShotList, err := st.ds.Get(datastore.NewKey(SnapShotList))
	if err == nil && snapShotList != nil {
		ssCidList := new(SnapShotCidList)
		err := json.Unmarshal(snapShotList, ssCidList)
		if err != nil {
			return fmt.Errorf("failed to load the snapshot cid list from datastore")
		}

		*ssCidList = append(*ssCidList, newSsCid)
		ssCidListBytes, err := json.Marshal(ssCidList)
		if err != nil {
			return fmt.Errorf("failed to marshal the cidlist of snapshot. %s", err.Error())
		}
		err = st.ds.Put(datastore.NewKey(SnapShotList), ssCidListBytes)
		if err != nil {
			return fmt.Errorf("failed to save the new snap shot cid list in ds")
		}
	} else {
		ssCidList := &SnapShotCidList{newSsCid}
		ssCidListBytes, err := json.Marshal(ssCidList)
		if err != nil {
			return fmt.Errorf("failed to marshal the cidlist of snapshot. %s", err.Error())
		}
		err = st.ds.Put(datastore.NewKey(SnapShotList), ssCidListBytes)
		if err != nil {
			return fmt.Errorf("failed to save the new snap shot cid list in ds")
		}
	}
	return nil
}

func (st *StateTree) GetSnapShotCidList() ([]cid.Cid, error) {
	snapShotList, err := st.ds.Get(datastore.NewKey(SnapShotList))
	if err == nil && snapShotList != nil {
		ssCidList := new(SnapShotCidList)
		err := json.Unmarshal(snapShotList, ssCidList)
		if err != nil {
			return nil, fmt.Errorf("failed to load the snapshot cid list from datastore")
		}
		return *ssCidList, nil
	}
	return nil, nil
}

func (st *StateTree) GetSnapShot(sscid cid.Cid) (shot *types.SnapShot, err error) {
	ss := new(types.SnapShot)
	err = st.Store.Get(context.Background(), sscid, ss)
	if err != nil {
		return nil, err
	}
	return ss, nil
}
