package option

const (
	defaultHttpAPIListenAddress = "/ip4/0.0.0.0/tcp/9000"
	defaultGraphqlListenAddress = "/ip4/0.0.0.0/tcp/9001"

	defaultDisableP2P = false
	defaultP2PAddress = "/ip4/0.0.0.0/tcp/8000"
)

type ServerAddress struct {
	HttpAPIListenAddress string `yaml:"HttpAPIListenAddress"`
	GraphqlListenAddress string `yaml:"GraphqlListenAddress"`

	DisableP2P bool   `yaml:"DisableP2P"`
	P2PAddress string `yaml:"P2PAddress"`
}
