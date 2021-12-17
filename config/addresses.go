package config

const (
	defaultP2PAddr = "/ip4/0.0.0.0/tcp/9000"
	defaultAdmin   = "/ip4/0.0.0.0/tcp/9001"
	defaultPando   = "/ip4/0.0.0.0/tcp/9002"
)

// Addresses stores the (string) multiaddr addresses for the node.
type Addresses struct {
	// AdminServer is the admin http server listen address
	AdminServer string
	// PandoServer is the Pando http server listen address
	PandoServer string
	// DisbleP2P disables libp2p hosting
	DisableP2P bool
	// P2PMaddr is the libp2p host multiaddr for all servers
	P2PAddr string
}
