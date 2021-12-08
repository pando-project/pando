package legs

import (
	"Pando/internal/metrics"
	"context"
	"fmt"
	dt "github.com/filecoin-project/go-data-transfer"
	"github.com/ipfs/go-cid"
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
	fmt.Printf("[go-legs onEvent] transfer event: %d, cid: %s\n", event.Code, channelState.BaseCID())
	if event.Code == dt.Open {
		recorder.lock.Lock()
		recorder.record[channelState.BaseCID()] =
			metrics.APITimer(context.Background(), metrics.GraphPersistenceLatency)
		recorder.lock.Unlock()
		fmt.Printf("[go-legs onEvent] start time recorded")
	}
	if event.Code == dt.FinishTransfer {
		recorder.lock.Lock()
		if record, exist := recorder.record[channelState.BaseCID()]; exist {
			record()
		}
		recorder.lock.Unlock()
	}
}
