package config

const (
	defaultGraphSync = "/ip4/127.0.0.1/tcp/3002"
	defaultP2PAddr   = "/ip4/0.0.0.0/tcp/3003"
)

// Addresses stores the (string) multiaddr addresses for the node.
type Addresses struct {
	// Admin is the admin http listen address
	GraphSync string
	// DisbleP2P disables libp2p hosting
	DisableP2P bool
	// P2PMaddr is the libp2p host multiaddr for all servers
	P2PAddr string
}
