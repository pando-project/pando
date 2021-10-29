package legs

import (
	"context"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/libp2p/go-libp2p-core/host"
)

var log = logging.Logger("core")

type LegsCore struct {
	Host       *host.Host
	DS         datastore.Batching
	LinkSystem *ipld.LinkSystem
	subs       []golegs.LegSubscriber
}

func NewLegsCore(host *host.Host, ds datastore.Batching, linkSys *ipld.LinkSystem) (*LegsCore, error) {

	return &LegsCore{
		Host:       host,
		DS:         ds,
		LinkSystem: linkSys,
	}, nil
}

func (core *LegsCore) NewMultiSubscriber(topic string) (golegs.LegMultiSubscriber, error) {
	lms, err := golegs.NewMultiSubscriber(context.Background(), *core.Host, core.DS, *core.LinkSystem, topic)
	if err != nil {
		return nil, err
	}
	return lms, nil
}

func (core *LegsCore) NewSubscriber(topic string) (golegs.LegSubscriber, error) {
	ls, err := golegs.NewSubscriber(context.Background(), *core.Host, core.DS, *core.LinkSystem, topic)
	if err != nil {
		return nil, err
	}

	watcher, _ := ls.OnChange()
	go validateReceived(watcher, core.DS)
	core.subs = append(core.subs, ls)
	return ls, nil
}

func validateReceived(watcher chan cid.Cid, ds datastore.Batching) {
	for {
		select {
		case downstream := <-watcher:
			if v, err := ds.Get(datastore.NewKey(downstream.String())); err != nil {
				log.Error("data not in receiver store: %v", err)
			} else {
				log.Debugf("Received from graphsync:\r\n cid: %s len:%d\r\n value:%s", downstream.String(), len(v), v)
			}
		}
	}
}
