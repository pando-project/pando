package legs

import (
	"Pando/internal/account"
	"context"
	"github.com/ipfs/go-graphsync"
	"github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/time/rate"
	"time"
)

func (l *Core) rateLimitHook() graphsync.OnOutgoingRequestHook {
	return func(p peer.ID, request graphsync.RequestData, hookActions graphsync.OutgoingRequestHookActions) {
		accountInfo := account.FetchPeerType(p, l.rateLimiter.Config().Registry)
		peerRateLimiter := l.rateLimiter.PeerLimiter(p)
		if peerRateLimiter == nil {
			peerRateLimiter = l.addPeerLimiter(p, accountInfo.PeerType, accountInfo.AccountLevel)
			// when recognize an unregistered provider, lock the request forever
			if peerRateLimiter == nil {
				l.pauseRequest(request.ID())
			}
		}
		log.Debugf("rate limit for peer %s is %f token/s, accountLevel is %v", p, peerRateLimiter.Limit(), accountInfo.AccountLevel)
		if !l.rateLimiter.Allow() || !peerRateLimiter.Allow() {
			const limitError = "your request was paused because of the rate limit policy"
			//go l.cancelRequest(request.ID())
			go l.pauseRequest(request.ID())
			log.Warnf(limitError)
			go l.unpauseRequest(request.ID(), peerRateLimiter)
			log.Debugf("leave rateLimitHook")
			return
		}
		log.Debugf("request %d from peer %s allowed", request.ID(), p)
	}
}

func (l *Core) cancelRequest(request graphsync.RequestID) {
	if err := l.lms.GraphSync().CancelRequest(context.Background(), request); err != nil {
		log.Warnf("cancel request failed, error: %s", err.Error())
	}
	log.Debugf("request %d canceled", request)
}

func (l *Core) pauseRequest(request graphsync.RequestID) {
	if err := l.lms.GraphSync().PauseRequest(request); err != nil {
		log.Warnf("pause request failed, error: %s", err.Error())
	}
}

func (l *Core) unpauseRequest(request graphsync.RequestID, peerRateLimiter *rate.Limiter) {
	time.Sleep(time.Second)
	if l.rateLimiter.Allow() && peerRateLimiter.Allow() {
		if err := l.lms.GraphSync().UnpauseRequest(request); err != nil {
			log.Warnf("unpause request %d failed, error: %s", request, err.Error())
		} else {
			log.Debugf("request %d unpaused", request)
		}
	} else {
		l.unpauseRequest(request, peerRateLimiter)
	}
}

func (l *Core) addPeerLimiter(peerID peer.ID, peerType account.PeerType, accountLevel int) *rate.Limiter {
	const action = "add peer limiter"
	var limiter *rate.Limiter
	var err error
	baseTokenRate := l.rateLimiter.Config().BaseTokenRate
	switch peerType {
	case account.UnregisteredPeer:
		return nil
		//limiter, err = l.rateLimiter.UnregisteredLimiter(baseTokenRate)
		//checkError(action, err)
	case account.WhiteListPeer:
		limiter, err = l.rateLimiter.WhitelistLimiter(baseTokenRate)
		checkError(action, err)
	case account.RegisteredPeer:
		limiter, err = l.rateLimiter.RegisteredLimiter(baseTokenRate, accountLevel, l.rateLimiter.Config().Registry.AccountLevelCount())
		checkError(action, err)
	}

	return l.rateLimiter.AddPeerLimiter(peerID, limiter)
}

func checkError(action string, e error) {
	if e != nil {
		log.Errorf("%s failed, error: %v", action, e)
	}
}
