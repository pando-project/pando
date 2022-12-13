package registry

import (
	"context"
	"github.com/ipfs/go-cid"
	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pando-project/pando/pkg/option"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

var MockAclCfg = option.AccountLevel{Threshold: []int{1, 10, 99}}
var trustedID = "12D3KooWK7CTS7cyWi51PeNE3cTjS2F2kDCZaQVU4A5xBmb9J1do"

func TestPollProvider(t *testing.T) {
	Convey("test poll provider", t, func() {
		cfg := &option.Discovery{
			Policy: option.Policy{
				Allow: true,
				Trust: true,
			},
			RediscoverWait: option.Duration(time.Minute).String(),
		}

		ctx := context.Background()
		// Create datastore
		dstore, err := leveldb.NewDatastore(t.TempDir(), nil)
		if err != nil {
			t.Fatal(err)
		}
		r, err := NewRegistry(ctx, cfg, &MockAclCfg, dstore, nil)
		if err != nil {
			t.Fatal(err)
		}

		peerID, err := peer.Decode(trustedID)
		if err != nil {
			t.Fatal("bad provider ID:", err)
		}

		err = r.RegisterOrUpdate(ctx, peerID, cid.Undef, peerID, cid.Undef, true)
		if err != nil {
			t.Fatal("failed to register directly:", err)
		}

		stopAfter := time.Hour

		// Check for auto-sync after pollInterval 0.
		r.pollProviders(0, stopAfter)
		timeout := time.After(2 * time.Second)
		select {
		case <-r.SyncChan():
		case <-timeout:
			t.Fatal("Expected sync channel to be written")
		}

		// Check that actions chan is not blocked by unread auto-sync channel.
		r.pollProviders(0, stopAfter)
		r.pollProviders(0, stopAfter)
		r.pollProviders(0, stopAfter)
		done := make(chan struct{})
		r.actions <- func() {
			close(done)
		}
		select {
		case <-done:
		case <-timeout:
			t.Fatal("actions channel blocked")
		}
		select {
		case <-r.SyncChan():
		case <-timeout:
			t.Fatal("Expected sync channel to be written")
		}

		// Set stopAfter to 0 so that stopAfter will have elapsed since last
		// contact. This will make publisher appear unresponsive and polling will
		// stop.
		stopAfter = 0
		r.pollProviders(0, stopAfter)
		r.pollProviders(0, stopAfter)
		r.pollProviders(0, stopAfter)

		// Check that publisher has been removed from provider info when publisher
		// appeared non-responsive.
		pinfo := r.ProviderInfo(peerID)
		if pinfo == nil {
			t.Fatal("did not find registered provider")
		}
		if err = pinfo[0].Publisher.Validate(); err == nil {
			t.Fatal("should not have valid publisher after polling stopped")
		}

		// Check that sync channel was not written since polling should have
		// stopped.
		select {
		case <-r.SyncChan():
			t.Fatal("sync channel should not have beem written to")
		default:
		}

		err = r.Close()
		if err != nil {
			t.Fatal(err)
		}
	})
}
