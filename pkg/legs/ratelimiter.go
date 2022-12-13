package legs

import (
	"context"
	"github.com/ipfs/go-graphsync"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pando-project/pando/pkg/account"
	"golang.org/x/time/rate"
	"time"
)

func (c *Core) rateLimitHook() graphsync.OnOutgoingRequestHook {
	return func(p peer.ID, request graphsync.RequestData, hookActions graphsync.OutgoingRequestHookActions) {
		accountInfo := account.FetchPeerType(p, c.rateLimiter.Config().Registry)
		peerRateLimiter := c.rateLimiter.PeerLimiter(p)
		if peerRateLimiter == nil {
			peerRateLimiter = c.addPeerLimiter(p, accountInfo.PeerType, accountInfo.AccountLevel)
		}
		logger.Debugf("rate limit for peer %s is %f token/s, accountLevel is %v",
			p, peerRateLimiter.Limit(), accountInfo.AccountLevel)
		if !c.rateLimiter.Allow() || !peerRateLimiter.Allow() {
			const limitError = "your request was paused because of the rate limit policy"
			go c.pauseRequest(request.ID())
			logger.Warnf(limitError)
			go c.unpauseRequest(request.ID(), peerRateLimiter)
			logger.Debugf("leave rateLimitHook")
			return
		}
		logger.Debugf("request %d from peer %s allowed", request.ID(), p)
	}
}

func (c *Core) pauseRequest(request graphsync.RequestID) {
	if err := c.GS.Pause(context.Background(), request); err != nil {
		logger.Warnf("pause request failed, error: %s", err.Error())
	}
}

func (c *Core) unpauseRequest(request graphsync.RequestID, peerRateLimiter *rate.Limiter) {
	time.Sleep(time.Second)
	if c.rateLimiter.Allow() && peerRateLimiter.Allow() {
		if err := c.GS.Unpause(context.Background(), request); err != nil {
			logger.Warnf("unpause request %d failed, error: %s", request, err.Error())
		} else {
			logger.Debugf("request %d unpaused", request)
		}
	} else {
		c.unpauseRequest(request, peerRateLimiter)
	}
}

func (c *Core) addPeerLimiter(peerID peer.ID, peerType account.PeerType, accountLevel int) *rate.Limiter {
	const action = "add peer limiter"
	var limiter *rate.Limiter
	var err error
	baseTokenRate := c.rateLimiter.Config().BaseTokenRate
	switch peerType {
	case account.UnregisteredPeer:
		limiter, err = c.rateLimiter.UnregisteredLimiter(baseTokenRate)
		checkError(action, err)
	case account.WhiteListPeer:
		limiter, err = c.rateLimiter.WhitelistLimiter(baseTokenRate)
		checkError(action, err)
	case account.RegisteredPeer:
		limiter, err = c.rateLimiter.RegisteredLimiter(baseTokenRate, accountLevel, c.rateLimiter.Config().Registry.AccountLevelCount())
		checkError(action, err)
	}

	return c.rateLimiter.AddPeerLimiter(peerID, limiter)
}

func checkError(action string, e error) {
	if e != nil {
		logger.Errorf("%s failed, error: %v", action, e)
	}
}
