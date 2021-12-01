package mock

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"testing"
)

func TestMockDiscoverer_Discover(t *testing.T) {
	m, e := newMockDiscoverer("12D3KooWQmrFvr4M7kbicnKbUkwuaNYJAfvKH5jiGj3vzmFW5n3E")
	if e != nil {
		t.Errorf(e.Error())
	}
	p, _ := peer.Decode("12D3KooWQmrFvr4M7kbicnKbUkwuaNYJAfvKH5jiGj3vzmFW5n3E")
	d, e := m.Discover(context.Background(), p, "bad1234")
	if e != nil {
		t.Errorf(e.Error())
	} else {
		fmt.Printf("type: %d, addr info: %s", d.Type, d.AddrInfo)
	}
}
