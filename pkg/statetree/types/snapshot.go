package types

import "fmt"

//go:generate cbor-gen-for ExtraInfo SnapShot
type SnapShot struct {
	Update       map[string]*ProviderState
	Height       uint64
	CreateTime   uint64
	PrevSnapShot string
	ExtraInfo    *ExtraInfo
}

type ExtraInfo struct {
	PeerID         string
	MultiAddresses string
}

func (t *ExtraInfo) String() string {
	str := ""
	str += fmt.Sprintf("peer id: %s, multiaddresses: %s ", t.PeerID, t.MultiAddresses)
	return str
}
