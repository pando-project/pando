package metadata

import (
	"Pando/statetree/types"
	"context"
	"errors"
	"fmt"
	dag "github.com/ipfs/go-merkledag"
	"path"

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

var (
	SnapShotDuration       = time.Second * 5
	BackupMaxInterval      = time.Second * 3
	BackupMaxDagNums       = 10000
	BackupTmpDir           = "./tmp/"
	BackFileName           = "backup.car"
	BackupCheckNumInterval = time.Second * 60
)

var NoRecordBackup = errors.New("no records need backup")

type MetaManager struct {
	flushTime         time.Duration
	recvCh            chan *MetaRecord
	outStateTreeCh    chan map[peer.ID]*types.ProviderState
	backupCh          chan cid.Cid
	ds                datastore.Datastore
	bs                blockstore.Blockstore
	dagServ           format.NodeGetter
	cache             map[peer.ID][]*MetaRecord
	mutex             sync.Mutex
	backupMaxInterval time.Duration
}

type backupRecord struct {
	cid      cid.Cid
	isBackup bool
}

type MetaRecord struct {
	Cid        cid.Cid
	ProviderID peer.ID
	Time       uint64
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore) (*MetaManager, error) {
	mm := &MetaManager{
		flushTime:         SnapShotDuration,
		recvCh:            make(chan *MetaRecord),
		outStateTreeCh:    make(chan map[peer.ID]*types.ProviderState),
		backupCh:          make(chan cid.Cid, 1000),
		ds:                ds,
		bs:                bs,
		cache:             make(map[peer.ID][]*MetaRecord),
		dagServ:           dag.NewDAGService(bsrv.New(bs, offline.Exchange(bs))),
		backupMaxInterval: BackupMaxInterval,
	}

	go mm.dealReceivedMeta()
	go mm.flushRegular()
	go mm.backupDagToFilecoin(ctx)
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
			mm.backupCh <- r.Cid
		}
		mm.mutex.Unlock()
	}
}

func (mm *MetaManager) flushRegular() {

	for range time.NewTicker(mm.flushTime).C {
		update := make(map[peer.ID]*types.ProviderState)
		for peerID, records := range mm.cache {
			cidlist := make([]cid.Cid, 0)
			for _, r := range records {
				cidlist = append(cidlist, r.Cid)
			}
			update[peerID] = &types.ProviderState{Cidlist: cidlist}
		}
		if len(update) > 0 {
			log.Debugw("send update to state tree")
			mm.outStateTreeCh <- update
		}
		// todo update state...hamt....

		mm.cache = make(map[peer.ID][]*MetaRecord)

	}
}

func (mm *MetaManager) GetMetaInCh() chan<- *MetaRecord {
	return mm.recvCh
}

func (mm *MetaManager) GetUpdateOut() <-chan map[peer.ID]*types.ProviderState {
	return mm.outStateTreeCh
}

func (mm *MetaManager) backupDagToFilecoin(ctx context.Context) {
	backupMutex := new(sync.Mutex)
	waitBackupRecoed := make([]*backupRecord, 0)
	t := time.NewTicker(time.Minute)
	go func() {
		for c := range mm.backupCh {
			log.Debugw("received dag to backup")
			waitBackupRecoed = append(waitBackupRecoed, &backupRecord{
				cid:      c,
				isBackup: false,
			})
			// delete record had been backup
			select {
			case _ = <-t.C:
				for i, r := range waitBackupRecoed {
					if r.isBackup {
						waitBackupRecoed = append(waitBackupRecoed[:i], waitBackupRecoed[i+1:]...)
					}
				}
			default:
			}
		}
	}()

	go func() {
		for range time.NewTicker(mm.backupMaxInterval).C {
			backupMutex.Lock()
			log.Infow("start backup the car to filecoin")
			// for update the isBackup later because the original slice has changed
			_waitBackupRecoed := make([]*backupRecord, len(waitBackupRecoed))
			copy(_waitBackupRecoed, waitBackupRecoed)
			err := mm.backupRecordsAndUpdateStatus(ctx, _waitBackupRecoed)
			if err != nil {
				continue
			}
			log.Infow("finished backup the car to filecoin")
			backupMutex.Unlock()
		}
	}()

	go func() {
		for range time.NewTicker(BackupCheckNumInterval).C {
			backupMutex.Lock()
			nums := 0
			for _, r := range waitBackupRecoed {
				if !r.isBackup {
					nums++
				}
			}
			if nums > BackupMaxDagNums {
				_waitBackupRecoed := make([]*backupRecord, len(waitBackupRecoed))
				copy(_waitBackupRecoed, waitBackupRecoed)
				err := mm.backupRecordsAndUpdateStatus(ctx, _waitBackupRecoed)
				if err != nil {
					continue
				}
				log.Infow("finished backup the car to filecoin")
			}
			backupMutex.Unlock()
		}
	}()
}

func (mm *MetaManager) backupRecordsAndUpdateStatus(ctx context.Context, _waitBackupRecoed []*backupRecord) error {
	waitBackupCidList := make([]cid.Cid, 0)
	for _, r := range _waitBackupRecoed {
		if !r.isBackup && r.cid != cid.Undef {
			waitBackupCidList = append(waitBackupCidList, r.cid)
		}
	}
	if len(waitBackupCidList) == 0 {
		log.Infow(NoRecordBackup.Error())
		return NoRecordBackup
	}
	err := ExportMetaCar(mm.dagServ, waitBackupCidList, BackupTmpDir+BackFileName, mm.bs)
	if err != nil {
		log.Errorf("failed to write Dags to car, err: %s", err.Error())
		return err
	}
	err = mm.makeBackupDeal(ctx, BackupTmpDir+BackFileName)
	if err != nil {
		log.Errorf("failed to backup car to filecoin, err: %s", err.Error())
		return err
	}
	for _, r := range _waitBackupRecoed {
		r.isBackup = true
	}
	return nil
}

func (mm *MetaManager) makeBackupDeal(ctx context.Context, path string) error {
	// todo
	return nil
}

func ExportMetaCar(dagds format.NodeGetter, cidlist []cid.Cid, filepath string, bs blockstore.Blockstore) error {
	_, err := os.Stat(path.Dir(filepath))
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path.Dir(filepath), os.ModePerm)
			if err != nil {
				log.Errorf("failed to create backup dir:%s , err:%s", path.Dir(filepath), err.Error())
			}
		} else {
			log.Errorf("please input correct filepath, err : %s", err.Error())
		}
	}
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("open file error : %s", err.Error())
		return err
	}
	defer f.Close()

	//todo debug
	for _, c := range cidlist {
		fmt.Println("cid: ", c)
		vb, e := bs.Get(c)
		if e != nil {
			return e
		}
		log.Debugf("[block] block value: ", vb.RawData())

		v, e := dagds.Get(context.Background(), c)
		if e != nil {
			return e
		}
		log.Debugf("[dag] block value: ", v.String())
	}

	err = car.WriteCar(context.Background(), dagds, cidlist, f)
	if err != nil {
		log.Errorf("failed to export the car for metadata, %s", err.Error())
		return err
	}
	return nil
}
