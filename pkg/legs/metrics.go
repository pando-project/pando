package legs

import (
	"context"
	dt "github.com/filecoin-project/go-data-transfer"
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando/pkg/metrics"
	"sync"
)

type dataTransferRecorder struct {
	record map[cid.Cid]func()
	lock   sync.Mutex
}

var recorder = dataTransferRecorder{
	record: make(map[cid.Cid]func()),
	lock:   sync.Mutex{},
}

func onDataTransferComplete(event dt.Event, channelState dt.ChannelState) {
	log.Debugf("transfer event: %s, cid: %s\n", dt.Events[event.Code], channelState.BaseCID())
	if event.Code == dt.Open {
		recorder.lock.Lock()
		recorder.record[channelState.BaseCID()] =
			metrics.APITimer(context.Background(), metrics.GraphPersistenceLatency)
		recorder.lock.Unlock()
	}
	if event.Code == dt.FinishTransfer {
		recorder.lock.Lock()
		if record, exist := recorder.record[channelState.BaseCID()]; exist {
			record()
		}
		recorder.lock.Unlock()
	}
}
