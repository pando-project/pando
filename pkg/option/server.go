package option

const (
	defaultAdminListenAddress   = "/ip4/127.0.0.1/tcp/8999"
	defaultHttpAPIListenAddress = "/ip4/0.0.0.0/tcp/9000"
	defaultGraphqlListenAddress = "/ip4/0.0.0.0/tcp/9001"

	defaultDisableP2P = false
	defaultP2PAddress = "/ip4/0.0.0.0/tcp/9002"

	defaultProfileListenAddress = "/ip4/0.0.0.0/tcp/9010"
)

type ServerAddress struct {
	AdminListenAddress   string `yaml:"AdminListenAddress"`
	HttpAPIListenAddress string `yaml:"HttpAPIListenAddress"`
	GraphqlListenAddress string `yaml:"GraphqlListenAddress"`

	DisableP2P bool   `yaml:"DisableP2P"`
	P2PAddress string `yaml:"P2PAddress"`

	ProfileListenAddress string `yaml:"ProfileListenAddress"`

	ExternalIP string `yaml:"ExternalIP"`
}
