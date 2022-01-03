package statetree

import (
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
	"pando/pkg/statetree/hamt"
	types2 "pando/pkg/statetree/types"
	"sync"
	"time"
)

const (
	// RootKey saves the hamt root cid in ds
	RootKey = "PandoStateTree"
	// SnapShotList saves the snapshots' cids in ds (bytes of []cid.Cid , not save cid of cidlist)
	SnapShotList = "PandoSnapShotList"
	// Reinitialize means if failed to load old tree, will initialize again
	Reinitialize = true
)

var log = logging.Logger("state-tree")

type SnapShotCidList []cid.Cid

type StateTree struct {
	ds    datastore.Batching
	Store cbor.IpldStore
	root  hamt.Map
	// the height of next snapshot. eg: height is 0 after initializing
	height       uint64
	snapShot     cid.Cid
	exinfo       *types2.ExtraInfo
	recvUpdateCh <-chan map[peer.ID]*types2.ProviderState
	ctx          context.Context
	cncl         func()
	// lock while updating hamt
	mtx sync.Mutex
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore, updateCh <-chan map[peer.ID]*types2.ProviderState, exinfo *types2.ExtraInfo) (*StateTree, error) {
	childCtx, cncl := context.WithCancel(ctx)
	cs := cbor.NewCborStore(bs)
	store := adt.WrapStore(childCtx, cs)
	st := &StateTree{
		exinfo:       exinfo,
		ds:           ds,
		Store:        store,
		recvUpdateCh: updateCh,
		ctx:          childCtx,
		cncl:         cncl,
	}
	// get the latest root(cid) from ds
	root, err := ds.Get(datastore.NewKey(RootKey))
	if err == nil && root != nil {
		_, rootcid, err := cid.CidFromBytes(root)
		if err != nil {
			return nil, fmt.Errorf("failed to load the State root from datastore")
		}
		log.Debugf("find root cid %s, loading...", rootcid.String())

		m, err := adt.AsMap(store, rootcid, builtin.DefaultHamtBitwidth)
		if err != nil && !Reinitialize {
			return nil, fmt.Errorf("failed to create hamt root by cid: %s\r\n%s", rootcid.String(), err.Error())
		} else if err != nil {
			emptyRoot, err := adt.MakeEmptyMap(store, builtin.DefaultHamtBitwidth)
			if err != nil {
				return nil, err
			}
			st.root = emptyRoot
		} else {
			st.root = m
		}
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
		ss := new(types2.SnapShot)
		newestSsCid := snapShotCidList[len(snapShotCidList)-1]
		err = store.Get(childCtx, newestSsCid, ss)
		if err != nil {
			return nil, fmt.Errorf("failed to load the newest snapshot: %s", newestSsCid.String())
		}
		st.snapShot = newestSsCid
		st.height = uint64(len(snapShotCidList))
	} else {
		st.snapShot = cid.Undef
		st.height = 0
	}

	// if MetadataManager send metadata update, update the State, save in hamt and change the root cid
	go st.Update(childCtx)

	return st, nil
}

func (st *StateTree) Update(ctx context.Context) {
	for {
		select {
		case _ = <-ctx.Done():
			log.Warn(ctx.Err().Error())
			return
		case update, ok := <-st.recvUpdateCh:
			if !ok {
				log.Error("metadata manager close the update channel.")
				st.cncl()
				log.Error("exit the state tree")
				return
			}
			rootCid, err := st.UpdateRoot(ctx, update)
			if err != nil {
				log.Errorf("while updating the state tree, some errors happened!\r\n%s", err.Error())
				st.cncl()
			}
			err = st.CreateSnapShot(ctx, rootCid, update)
			if err != nil {
				log.Errorf("while creating and saving snapshot, some errors happened!\r\n%s", err.Error())
				st.cncl()
			}
			st.height += 1
		}
	}
}

func (st *StateTree) UpdateRoot(ctx context.Context, update map[peer.ID]*types2.ProviderState) (cid.Cid, error) {
	log.Debug("start updating the state tree")
	st.mtx.Lock()
	defer st.mtx.Unlock()
	for p, cidlist := range update {
		state := new(types2.ProviderState)
		found, err := st.root.Get(hamt.ProviderKey{ID: p}, state)
		if err != nil {
			return cid.Undef, fmt.Errorf("failed to get provider state from hamt, %s", err.Error())
		}
		if !found {
			err = st.root.Put(hamt.ProviderKey{ID: p}, &types2.ProviderState{Cidlist: cidlist.Cidlist, LastCommitHeight: st.height})
			if err != nil {
				return cid.Undef, fmt.Errorf("failed to put provider state into hamt, %s", err.Error())
			}
		} else {
			state.Cidlist = append(state.Cidlist, cidlist.Cidlist...)
			state.LastCommitHeight = st.height
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

func (st *StateTree) CreateSnapShot(ctx context.Context, newRoot cid.Cid, update map[peer.ID]*types2.ProviderState) error {
	var height uint64
	var previousSs string
	if st.snapShot == cid.Undef {
		// todo
		height = uint64(0)
		previousSs = ""
	} else {
		//oldSs := new(types.SnapShot)
		//err := st.Store.Get(ctx, st.snapShot, oldSs)
		//if err != nil {
		//	return fmt.Errorf("failed to load the old snapshot root from datastore")
		//}
		//height = oldSs.Height + 1
		height = st.height
		previousSs = st.snapShot.String()
	}

	_update := make(map[string]*types2.ProviderState)
	for p, state := range update {
		_update[p.String()] = state
		// there is no height info from metamanager
		_update[p.String()].LastCommitHeight = st.height
	}

	newSs := &types2.SnapShot{
		Update:       _update,
		Height:       height,
		CreateTime:   uint64(time.Now().UnixNano()),
		PrevSnapShot: previousSs,
		ExtraInfo:    st.exinfo,
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

func (st *StateTree) GetSnapShot(sscid cid.Cid) (shot *types2.SnapShot, err error) {
	ss := new(types2.SnapShot)
	err = st.Store.Get(st.ctx, sscid, ss)
	if err == blockstore.ErrNotFound {
		return nil, NotFoundErr
	} else if err != nil {
		return nil, err
	}
	return ss, nil
}

func (st *StateTree) GetSnapShotByHeight(height uint64) (*types2.SnapShot, error) {
	if height < 0 {
		return nil, fmt.Errorf("height must be positive")
	}
	cidlist, err := st.GetSnapShotCidList()
	if err != nil {
		return nil, err
	}
	if len(cidlist) == 0 || height > uint64(len(cidlist)-1) {
		log.Warnf("height cannot be bigger than newest")
		return nil, NotFoundErr
	}
	ss, err := st.GetSnapShot(cidlist[height])
	if err != nil {
		return nil, err
	}
	return ss, nil
}

// GetProviderStateByPeerID for graphql
func (st *StateTree) GetProviderStateByPeerID(id peer.ID) (*types2.ProviderStateRes, error) {
	st.mtx.Lock()
	defer st.mtx.Unlock()
	state := new(types2.ProviderState)
	found, err := st.root.Get(hamt.ProviderKey{ID: id}, state)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider state from hamt, %s", err.Error())
	}

	if !found {
		return nil, NotFoundErr
	} else {
		res := new(types2.ProviderStateRes)
		lastUpdateHeight := state.LastCommitHeight
		cidlist, err := st.GetSnapShotCidList()
		if err != nil {
			return nil, fmt.Errorf("failed to get snapshot cidlist")
		}
		sscid := cidlist[lastUpdateHeight]
		ss, err := st.GetSnapShot(sscid)
		if err != nil {
			return nil, fmt.Errorf("failed to get snapshot")
		}
		res.State = *state
		res.NewestUpdate = ss.Update[id.String()].Cidlist
		return res, nil
	}
}

func (st *StateTree) DeleteInfo() error {
	err := st.ds.Delete(datastore.NewKey(RootKey))
	if err != nil {
		return err
	}
	err = st.ds.Delete(datastore.NewKey(SnapShotList))
	if err != nil {
		return err
	}
	return nil
}

func (st *StateTree) GetPandoInfo() (*types2.ExtraInfo, error) {
	if st.exinfo != nil {
		return st.exinfo, nil
	}
	return nil, fmt.Errorf("nil info")
}

func (st *StateTree) Shutdown() error {
	st.cncl()
	log.Warn("shutting down the state tree...")
	return nil
}