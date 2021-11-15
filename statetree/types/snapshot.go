package types

//go:generate cbor-gen-for ExtraInfo SnapShot
type SnapShot struct {
	Update       map[string]*ProviderState
	Height       uint64
	CreateTime   uint64
	PrevSnapShot string
	ExtraInfo    *ExtraInfo
}

type ExtraInfo struct {
	GraphSyncUrl   string
	GoLegsSubUrl   string
	GolegsSubTopic string
	MultiAddr      string
}
