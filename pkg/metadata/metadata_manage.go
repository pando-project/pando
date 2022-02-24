package metadata

import (
	"context"
	"fmt"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	selectorparse "github.com/ipld/go-ipld-prime/traversal/selector/parse"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/pkg/registry"
	"github.com/kenlabs/pando/pkg/statetree/types"
	"github.com/libp2p/go-libp2p-core/peer"
	"os"
	"path"

	"sync"
	"time"
)

var log = logging.Logger("meta-manager")

var (
	SnapShotDuration = time.Second * 5
	BackupTmpDirName = "ttmp"
	BackupTmpPath    string
	BackFileName     = "backup-%s-%d.car"
	syncPrefix       = "/sync/"
)

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	BackupTmpDir := path.Join(pwd, BackupTmpDirName)
	_, err = os.Stat(BackupTmpDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(BackupTmpDir, os.ModePerm)
			if err != nil {
				log.Errorf("failed to create backup dir:%s , err:%s", BackupTmpDir, err.Error())
			}
		} else {
			log.Errorf("please input correct filepath, err : %s", err.Error())
		}
	}
	BackupTmpPath = BackupTmpDir
}

type MetaManager struct {
	flushTime         time.Duration
	recvCh            chan *MetaRecord
	outStateTreeCh    chan map[peer.ID]*types.ProviderState
	backupCh          chan cid.Cid
	ds                datastore.Datastore
	bs                blockstore.Blockstore
	cache             map[peer.ID][]*MetaRecord
	mutex             sync.Mutex
	backupMaxInterval time.Duration
	estBackupSys      *BackupSystem
	registry          *registry.Registry
	ls                *ipld.LinkSystem
	backupCfg         *option.Backup
	ctx               context.Context
	cncl              context.CancelFunc
}

type MetaRecord struct {
	Cid        cid.Cid
	ProviderID peer.ID
	Time       uint64
}

func New(ctx context.Context, ds datastore.Batching, bs blockstore.Blockstore, ls *ipld.LinkSystem, registry *registry.Registry, backupCfg *option.Backup) (*MetaManager, error) {
	ebs, err := NewBackupSys(backupCfg)
	if err != nil {
		return nil, err
	}

	cctx, cncl := context.WithCancel(ctx)

	mm := &MetaManager{
		flushTime:      SnapShotDuration,
		recvCh:         make(chan *MetaRecord),
		outStateTreeCh: make(chan map[peer.ID]*types.ProviderState),
		backupCh:       make(chan cid.Cid, 1000),
		ds:             ds,
		bs:             bs,
		ls:             ls,
		cache:          make(map[peer.ID][]*MetaRecord),
		estBackupSys:   ebs,
		registry:       registry,
		backupCfg:      backupCfg,
		ctx:            cctx,
		cncl:           cncl,
	}

	go mm.dealReceivedMeta()
	go mm.flushRegular()
	go mm.genCarForProviders(cctx)
	//go mm.backupDagToCarLocally(cctx)
	return mm, nil
}

func (mm *MetaManager) dealReceivedMeta() {
	for {
		select {
		case r, ok := <-mm.recvCh:
			if !ok {
				return
			}
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
}

func (mm *MetaManager) flushRegular() {

	for range time.NewTicker(mm.flushTime).C {
		select {
		case _ = <-mm.ctx.Done():
			return
		default:
		}
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
		mm.cache = make(map[peer.ID][]*MetaRecord)
	}
}

func (mm *MetaManager) GetMetaInCh() chan<- *MetaRecord {
	return mm.recvCh
}

func (mm *MetaManager) GetUpdateOut() <-chan map[peer.ID]*types.ProviderState {
	return mm.outStateTreeCh
}

func (mm *MetaManager) genCarForProviders(ctx context.Context) {
	interval, err := time.ParseDuration(mm.backupCfg.BackupGenInterval)
	if err != nil {
		log.Errorf("invalid BackupGenInterval config: %s\n err:%s\n",
			mm.backupCfg.BackupGenInterval, err.Error())
	}
	go func() {
		for range time.NewTicker(interval).C {
			select {
			case _ = <-ctx.Done():
				return
			default:
			}
			providersInfo := mm.registry.AllProviderInfo()
			for _, info := range providersInfo {
				lastBackup := info.LastBackupMeta
				lastSync, err := mm.ds.Get(ctx, datastore.NewKey(syncPrefix+info.AddrInfo.ID.String()))
				if err != nil {
					// register but not contact
					if err == datastore.ErrNotFound {
						continue
					}
					log.Errorf("failed to get last sync for provider:%s\n", info.AddrInfo.ID.String())
				}
				_, lastSyncCid, err := cid.CidFromBytes(lastSync)
				if err != nil {
					log.Errorf("failed to decode cid for last sync for provider:%s\nerr:%s\n",
						info.AddrInfo.ID.String(), err.Error())
				}
				if lastBackup == lastSyncCid {
					// not need back up
					continue
				}
				fname := fmt.Sprintf(BackFileName, info.AddrInfo.ID.String(), time.Now().UnixNano())
				err = mm.exportMetaCar(ctx, fname, lastSyncCid, lastBackup)
				if err != nil {
					log.Errorf("failed to export backup car for provider:%s\nerr:%s",
						info.AddrInfo.ID.String(), err.Error())
					continue
				}
				err = mm.registry.RegisterOrUpdate(ctx, info.AddrInfo.ID, lastSyncCid)
				if err != nil {
					log.Errorf("failed to update provider info for backup cid, err:%s\n", err.Error())
				}

			}
		}
	}()
}

func (mm *MetaManager) Close() {
	mm.cncl()
	close(mm.outStateTreeCh)
	close(mm.backupCh)
	close(mm.recvCh)
}

func (mm *MetaManager) exportMetaCar(ctx context.Context, filename string, root cid.Cid, lastBackup cid.Cid) error {
	f, err := os.OpenFile(path.Join(BackupTmpPath, filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("open file error : %s", err.Error())
		return err
	}
	defer func(f *os.File) {
		if f != nil {
			err := f.Close()
			if err != nil {
				log.Warnf("close car file failed, %v", err)
			}
		}
	}(f)
	var ss ipld.Node
	if !lastBackup.Equals(cid.Undef) {
		ss = golegs.ExploreRecursiveWithStopNode(selector.RecursionLimit{}, nil, cidlink.Link{lastBackup})
	} else {
		ss = selectorparse.CommonSelector_ExploreAllRecursively
	}

	_, err = car.TraverseV1(ctx, mm.ls, root, ss, f)
	if err != nil {
		log.Errorf("failed to export meta backup car, err:%s\n", err.Error())
		return err
	}
	return nil
}
