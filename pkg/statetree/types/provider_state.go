package types

import (
	"fmt"
	"github.com/ipfs/go-cid"
)

//go:generate cbor-gen-for ProviderState

type ProviderState struct {
	Cidlist          []cid.Cid
	LastCommitHeight uint64
}

// ProviderStateRes for graphql,
// include the total state about the provider and the newest state change(such as cidlist)
type ProviderStateRes struct {
	State        ProviderState
	NewestUpdate []cid.Cid
}

func (t *ProviderState) String() string {
	str := ""
	str += fmt.Sprintf("last Commit height: %d, ", t.LastCommitHeight)
	str += "cidlist: ["
	for _, c := range t.Cidlist {
		str += fmt.Sprintf("%s, ", c.String())
	}
	str += "]"
	return str
}
