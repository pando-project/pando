package legs

import (
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/time/rate"
	"pando/pkg/account"
	"pando/pkg/policy"
	"pando/pkg/registry"
	"reflect"
	"testing"
)

var testPeer, _ = peer.Decode("12D3KooWNtUworDmrdTUBrLqeD8s36MLnpRX1QJGQ46HXaJVBXV4")

func TestRateLimiterRank(t *testing.T) {
	Convey("when give different peerType then get different rateLimiter", t, func() {
		c := &Core{
			rateLimiter: &policy.Limiter{},
		}
		patch := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter), "Config", func(_ *policy.Limiter) policy.LimiterConfig {
			return policy.LimiterConfig{
				BaseTokenRate: 0,
				Registry:      &registry.Registry{},
			}
		})
		defer patch.Reset()
		patch2 := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter), "UnregisteredLimiter", func(_ *policy.Limiter, _ float64) (*rate.Limiter, error) {
			return nil, fmt.Errorf("unknown error")
		})
		defer patch2.Reset()
		//WhitelistLimiter
		patch3 := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter), "WhitelistLimiter", func(_ *policy.Limiter, _ float64) (*rate.Limiter, error) {
			return nil, fmt.Errorf("unknown error")
		})
		defer patch3.Reset()
		patch4 := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter), "RegisteredLimiter", func(_ *policy.Limiter, _ float64, _ int, _ int) (*rate.Limiter, error) {
			return nil, fmt.Errorf("unknown error")
		})
		defer patch4.Reset()
		patch5 := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter.Config().Registry), "AccountLevelCount", func(_ *registry.Registry) int {
			return -1
		})
		defer patch5.Reset()
		patch6 := gomonkey.ApplyMethod(reflect.TypeOf(c.rateLimiter), "AddPeerLimiter", func(_ *policy.Limiter, peerID peer.ID, limiter *rate.Limiter) *rate.Limiter {
			return nil
		})
		defer patch6.Reset()
		//(peerID peer.ID, limiter *rate.Limiter) *rate.Limiter

		c.addPeerLimiter(testPeer, account.RegisteredPeer, 0)
		c.addPeerLimiter(testPeer, account.UnregisteredPeer, 0)
		c.addPeerLimiter(testPeer, account.WhiteListPeer, 0)
	})
}
