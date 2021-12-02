package legs

import (
	"github.com/ipfs/go-graphsync"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (l *Core) recordIncomingResponseLatencyHook() graphsync.OnIncomingResponseHook {
	return func(p peer.ID, responseData graphsync.ResponseData, hookActions graphsync.IncomingResponseHookActions) {

	}
}
