package metadata

import (
	"context"
	"fmt"
	dag "github.com/ipfs/go-merkledag"

	bsrv "github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	format "github.com/ipfs/go-ipld-format"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-car"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"

	"sync"
	"time"
)

var log = logging.Logger("meta-manager")

const SnapShotDuration = time.Second * 5

type MetaManager struct {
	flushTime time.Duration
	recvCh    chan *MetaRecord
	ds        datastore.Datastore
	bs        blockstore.Blockstore
	dagds     format.NodeGetter
	cache     map[peer.ID][]*MetaRecord
	mutex     sync.Mutex
}

type MetaRecord struct {
	Cid        cid.Cid
	ProviderID peer.ID
	Time       uint64
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore) (*MetaManager, error) {
	mm := &MetaManager{
		flushTime: SnapShotDuration,
		recvCh:    make(chan *MetaRecord),
		ds:        ds,
		cache:     make(map[peer.ID][]*MetaRecord),
		dagds:     dag.NewDAGService(bsrv.New(bs, offline.Exchange(bs))),
	}

	go mm.dealReceivedMeta()
	go mm.flushRegular()

	return mm, nil
}

func (mm *MetaManager) dealReceivedMeta() {
	for r := range mm.recvCh {
		mm.mutex.Lock()
		if r != nil {
			if mm.cache[r.ProviderID] == nil {
				mm.cache[r.ProviderID] = make([]*MetaRecord, 0)
			}
			mm.cache[r.ProviderID] = append(mm.cache[r.ProviderID], r)
		}
		mm.mutex.Unlock()
	}
}

func (mm *MetaManager) flushRegular() {

	for range time.NewTicker(mm.flushTime).C {
		for peerID, records := range mm.cache {
			log.Debugf("write metadata to car for provider: %s", peerID.String())
			cidlist := make([]cid.Cid, 0)
			for _, r := range records {
				cidlist = append(cidlist, r.Cid)
			}
			exportMetaCar(mm.dagds, cidlist, "./received/"+peerID.String()[:5]+time.Now().String()+".car")
		}

		// todo update state...hamt....

		mm.cache = make(map[peer.ID][]*MetaRecord)

	}
}

func (mm *MetaManager) GetMetaInCh() chan<- *MetaRecord {
	return mm.recvCh
}

func exportMetaCar(dagds format.NodeGetter, cidlist []cid.Cid, path string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("open file error : %s", err.Error())
		return
	}
	defer f.Close()

	//todo debug
	for _, c := range cidlist {
		fmt.Println("cid: ", c)
		v, e := dagds.Get(context.Background(), c)
		if e != nil {
			fmt.Println(e)
		}
		fmt.Println("value: ", v.String())
	}

	err = car.WriteCar(context.Background(), dagds, cidlist, f)
	if err != nil {
		log.Errorf("failed to export the car for metadata, %s", err.Error())
	}

}
