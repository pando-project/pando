package config

import (
	"Pando/speedtester"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
)

type Option func(*Config)

func Init(out io.Writer, opts ...Option) (*Config, error) {
	identity, err := CreateIdentity(out)
	if err != nil {
		return nil, err
	}

	return InitWithIdentity(identity, opts...)
}

func InitWithIdentity(identity Identity, opts ...Option) (*Config, error) {
	conf := &Config{
		// setup the node's default addresses.
		Addresses: Addresses{
			GraphSync: defaultGraphSync,
			GraphQL:   defaultGraphQl,
			P2PAddr:   defaultP2PAddr,
			MetaData:  defaultMetaData,
			Admin:     defaultAdmin,
		},
		Identity: identity,
		Discovery: Discovery{
			LotusGateway: defaultLotusGateway,
			Policy: Policy{
				Allow: defaultAllow,
				Trust: defaultTrust,
			},
			PollInterval:   defaultPollInterval,
			RediscoverWait: defaultRediscoverWait,
			Timeout:        defaultDiscoveryTimeout,
		},

		Datastore: Datastore{
			Type: defaultDatastoreType,
			Dir:  defaultDatastoreDir,
		},
		AccountLevel:  AccountLevel{defaultThreshold},
		SingleDAGSize: defaultSingleDAGSize,
	}
	for _, opt := range opts {
		opt(conf)
	}

	// disable by option
	if conf.BandWidth != -1 {
		conf.BandWidth = speedtester.FetchInternetSpeed(false)
	} else {
		conf.BandWidth = 10
	}

	return conf, nil
}

// CreateIdentity initializes a new identity.
func CreateIdentity(out io.Writer) (Identity, error) {
	ident := Identity{}

	var sk crypto.PrivKey
	var pk crypto.PubKey

	_, _ = fmt.Fprintf(out, "generating ED25519 keypair...")
	priv, pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return ident, err
	}
	_, _ = fmt.Fprintf(out, "done\n")

	sk = priv
	pk = pub

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	skbytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)

	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	_, _ = fmt.Fprintf(out, "account identity: %s\n", ident.PeerID)
	return ident, nil
}
