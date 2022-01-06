package multiaddress

import (
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

func MultiaddressToNetAddress(multiAddrStr string) (string, error) {
	multiAddress, err := multiaddr.NewMultiaddr(multiAddrStr)
	if err != nil {
		return "", err
	}
	netAddress, err := manet.ToNetAddr(multiAddress)
	if err != nil {
		return "", err
	}

	return netAddress.String(), nil
}
