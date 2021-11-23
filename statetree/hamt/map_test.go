package hamt

import (
	"Pando/statetree/types"
	"context"
	"github.com/stretchr/testify/assert"

	//"Pando/statetree/hamt"

	"fmt"
	"github.com/filecoin-project/specs-actors/v5/actors/builtin"
	"github.com/filecoin-project/specs-actors/v5/actors/util/adt"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	cbor "github.com/ipfs/go-ipld-cbor"
	"testing"
)

func TestMapSaveAndLoad(t *testing.T) {
	ds := datastore.NewMapDatastore()
	mds := dssync.MutexWrap(ds)
	bs := blockstore.NewBlockstore(mds)
	cs := cbor.NewCborStore(bs)
	store := adt.WrapStore(context.TODO(), cs)
	emptyRoot, err := adt.MakeEmptyMap(store, builtin.DefaultHamtBitwidth)
	if err != nil {
		t.Error(err)
	}
	state1 := new(types.ProviderState)
	state2 := new(types.ProviderState)

	testCid1, _ := cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfaa")
	testCid2, _ := cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfab")
	testCid3, _ := cid.Decode("bafy2bzaceamp42wmmgr2g2ymg46euououzfyck7szknvfacqscohrvaikwfac")

	state1.Cidlist = append(state1.Cidlist, testCid1, testCid2)
	state2.Cidlist = append(state2.Cidlist, testCid3)

	err = emptyRoot.Put(ProviderKey{ID: "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV6"}, state1)
	assert.NoError(t, err)

	err = emptyRoot.Put(ProviderKey{ID: "12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4"}, state2)
	assert.NoError(t, err)

	newRootCid, err := emptyRoot.Root()
	assert.NoError(t, err)

	fmt.Println(newRootCid.String())
	err = ds.Put(datastore.NewKey("TESTROOTKEY"), newRootCid.Bytes())
	assert.NoError(t, err)

	rootCidBytest, err := ds.Get(datastore.NewKey("TESTROOTKEY"))
	assert.NoError(t, err)

	_, rootcid, err := cid.CidFromBytes(rootCidBytest)
	assert.NoError(t, err)

	fmt.Println(rootcid.String())

	_, err = adt.AsMap(store, rootcid, builtin.DefaultHamtBitwidth)
	assert.NoError(t, err)

}
