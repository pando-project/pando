package config

const (
	defaultAdmin     = "/ip4/0.0.0.0/tcp/9001"
	defaultGraphSync = "/ip4/0.0.0.0/tcp/9002"
	defaultGraphQl   = "/ip4/0.0.0.0/tcp/9003"
	defaultMetaData  = "/ip4/0.0.0.0/tcp/9004"
	defaultP2PAddr   = "/ip4/0.0.0.0/tcp/9000"
)

// Addresses stores the (string) multiaddr addresses for the node.
type Addresses struct {
	// Admin is the admin http server listen address
	Admin string
	// MetaData is the state tree http listen address
	MetaData string
	// Admin is the admin http listen address
	GraphSync string
	// DisbleP2P disables libp2p hosting
	DisableP2P bool
	// P2PMaddr is the libp2p host multiaddr for all servers
	P2PAddr string
	// GraphQL address
	GraphQL string
}
