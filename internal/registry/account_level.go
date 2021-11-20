package registry

import (
	"fmt"
	"github.com/libp2p/go-libp2p-core/peer"
	"math/big"
)

var (
	attoFIl = big.NewInt(1)
	FIL     = big.NewInt(1).Exp(big.NewInt(10), big.NewInt(18), nil)
)

func (r *Registry) getAccountLevel(balance *big.Int) (int, error) {
	level := 0
	ok := big.NewInt(1).Div(balance, FIL).IsUint64()
	if !ok {
		return -1, fmt.Errorf("valid balance: %s", balance.String())
	}
	balanceFIL := big.NewInt(1).Div(balance, FIL).Uint64()
	for _, l := range r.accountLevel.Threshold {
		if balanceFIL < uint64(l) {
			break
		}
		level += 1
	}
	return level, nil
}

func (r *Registry) ProviderAccountLevel(provider peer.ID) (int, error) {
	info := r.ProviderInfo(provider)
	if info == nil {
		return -1, fmt.Errorf("not register provider")
	}
	return info.AccountLevel, nil
}

func (r *Registry) AccountLevelCount() int {
	return len(r.accountLevel.Threshold)
}
