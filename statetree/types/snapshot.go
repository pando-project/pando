package types

import (
	"github.com/ipfs/go-cid"
)

//go:generate cbor-gen-for ExtraInfo SnapShot
type SnapShot struct {
	Update       map[string]*ProviderState
	Height       uint64
	CreateTime   uint64
	PrevSnapShot cid.Cid
	ExtraInfo    *ExtraInfo
}

type ExtraInfo struct {
	GraphSyncUrl   string
	GoLegsSubUrl   string
	GolegsSubTopic string
	MultiAddr      string
}
