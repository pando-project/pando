package est_utils

import (
	//"github.com/application-research/filclient"
	"github.com/libp2p/go-libp2p-core/peer"

	//"github.com/application-research/filclient"
	"github.com/ipfs/go-cid"
	"time"
)

type AddResponse struct {
	Cid       string
	EstuaryId uint64
	Providers []string
}

type ContentStatus struct {
	Content struct {
		Id           int    `json:"id"`
		Cid          string `json:"cid"`
		Name         string `json:"name"`
		UserId       int    `json:"userId"`
		Description  string `json:"description"`
		Size         int    `json:"size"`
		Active       bool   `json:"active"`
		Offloaded    bool   `json:"offloaded"`
		Replication  int    `json:"replication"`
		AggregatedIn int    `json:"aggregatedIn"`
		Aggregate    bool   `json:"aggregate"`
		Pinning      bool   `json:"pinning"`
		PinMeta      string `json:"pinMeta"`
		Failed       bool   `json:"failed"`
		Location     string `json:"location"`
		DagSplit     bool   `json:"dagSplit"`
	} `json:"content"`
	Deals         []*DealStatus `json:"deals"`
	FailuresCount int           `json:"failuresCount"`
}

type DealStatus struct {
	Deal           contentDeal       `json:"deal"`
	TransferStatus *ChannelState     `json:"transfer"`
	OnChainState   *onChainDealState `json:"onChainState"`
}

type contentDeal struct {
	Content          uint      `json:"content" gorm:"index:,option:CONCURRENTLY"`
	PropCid          DbCID     `json:"propCid"`
	Miner            string    `json:"miner"`
	DealID           int64     `json:"dealId"`
	Failed           bool      `json:"failed"`
	Verified         bool      `json:"verified"`
	FailedAt         time.Time `json:"failedAt,omitempty"`
	DTChan           string    `json:"dtChan" gorm:"index"`
	TransferStarted  time.Time `json:"transferStarted"`
	TransferFinished time.Time `json:"transferFinished"`

	OnChainAt time.Time `json:"onChainAt"`
	SealedAt  time.Time `json:"sealedAt"`
}

type DbCID struct {
	CID cid.Cid
}

type onChainDealState struct {
	SectorStartEpoch uint64 `json:"sectorStartEpoch"`
	LastUpdatedEpoch uint64 `json:"lastUpdatedEpoch"`
	SlashEpoch       uint64 `json:"slashEpoch"`
}

type ChannelState struct {
	//datatransfer.Channel

	// SelfPeer returns the peer this channel belongs to
	SelfPeer   peer.ID `json:"selfPeer"`
	RemotePeer peer.ID `json:"remotePeer"`

	// Status is the current status of this channel
	Status    uint64 `json:"status"`
	StatusStr string `json:"statusMessage"`

	// Sent returns the number of bytes sent
	Sent uint64 `json:"sent"`

	// Received returns the number of bytes received
	Received uint64 `json:"received"`

	// Message offers additional information about the current status
	Message string `json:"message"`

	BaseCid string `json:"baseCid"`

	ChannelID interface{} `json:"channelId"`
}
