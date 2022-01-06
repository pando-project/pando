package provider

import ipldFormat "github.com/ipfs/go-ipld-format"

type Provider interface {
	ConnectPando(peerAddress string, peerID string) error
	Close() error
	Push(node ipldFormat.Node) error
}
