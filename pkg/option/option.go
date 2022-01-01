package option

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"pando/pkg/version"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Options is the server daemon start-up options
type Options struct {
	flags   *pflag.FlagSet `yaml:"-"`
	viper   *viper.Viper   `yaml:"-"`
	yamlStr string         `yaml:"-"`

	// Flags from command line only
	ShowVersion bool   `yaml:"-"`
	ShowHelp    bool   `yaml:"-"`
	ShowConfig  bool   `yaml:"-"`
	ConfigFile  string `yaml:"-"`

	PandoRoot        string `yaml:"PandoRoot"`
	DisableSpeedTest bool   `yaml:"DisableSpeedTest"`
	// Supported LogLevels are: DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL, and
	// their lower-case forms.
	LogLevel      string        `yaml:"LogLevel"`
	Identity      Identity      `yaml:"Identity"`
	ServerAddress ServerAddress `yaml:"ServerAddress"`
	DataStore     DataStore     `yaml:"DataStore"`
	Discovery     Discovery     `yaml:"Discovery"`
	AccountLevel  AccountLevel  `yaml:"AccountLevel"`
	RateLimit     RateLimit     `yaml:"RateLimit"`
	Backup        Backup        `yaml:"Backup"`
}

// New creates a default Options.
func New(root *cobra.Command) *Options {
	var opt *Options
	if root == nil {
		opt = &Options{
			flags: pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError),
			viper: viper.New(),
		}
	} else {
		opt = &Options{
			flags: root.PersistentFlags(),
			viper: viper.New(),
		}
	}

	// options for root and daemon
	opt.flags.BoolVarP(&opt.ShowVersion, "version", "v", false,
		"Print the version.")

	opt.flags.BoolVarP(&opt.ShowHelp, "help", "h", false,
		"Print the helper info.")

	opt.flags.BoolVarP(&opt.ShowConfig, "print-config", "c", false,
		"Print the configuration info.")

	opt.flags.StringVarP(&opt.ConfigFile, "config-file", "f", "config.yaml",
		"Load server configuration from a yaml file rather than line flags.")

	homeRoot, err := homedir.Expand("~/.pando")
	if err != nil {
		panic(fmt.Errorf("expand home dir failed when new options: %v", err))
	}
	opt.flags.StringVarP(&opt.PandoRoot, "pando-root", "r", homeRoot,
		"Pando root directory which stores data and config file.")

	opt.flags.BoolVarP(&opt.DisableSpeedTest, "disable-speedtest", "p", false,
		"Init server configuration without Internet connection speed test.")

	opt.flags.StringVar(&opt.LogLevel, "log-level", "debug",
		"Log level of Pando.")

	// options for server address
	opt.flags.StringVar(&opt.ServerAddress.HttpAPIListenAddress, "http-listen-addr", defaultHttpAPIListenAddress,
		fmt.Sprintf("Http server listen address(in multiaddress format, like %s).", defaultHttpAPIListenAddress))

	opt.flags.StringVar(&opt.ServerAddress.GraphqlListenAddress, "graphql-listen-addr", defaultGraphqlListenAddress,
		fmt.Sprintf("Graphql server listen address(in multiaddress format, like %s).", defaultGraphqlListenAddress))

	opt.flags.BoolVar(&opt.ServerAddress.DisableP2P, "p2p-disable", defaultDisableP2P,
		"Disable libp2p hosting.")

	opt.flags.StringVar(&opt.ServerAddress.P2PAddress, "p2p-address", defaultP2PAddress,
		fmt.Sprintf("P2P hosting address(in multiaddress format, like %s).", defaultP2PAddress))

	// options for datastore
	opt.flags.StringVar(&opt.DataStore.Type, "datastore-type", defaultDataStoreType,
		"Datastore type, support levelds only for now.")

	opt.flags.StringVar(&opt.DataStore.Dir, "datastore-dir", defaultDataStoreDir,
		"Directory of datastore files.")

	// options for discovery
	opt.flags.StringVar(&opt.Discovery.LotusGateway, "discovery-lotus-gateway", defaultLotusGateway,
		"Lotus gateway address.")

	opt.flags.BoolVar(&opt.Discovery.Policy.Allow, "discovery-policy-allow", defaultAllow,
		"Discovery allow all origin source host.")

	opt.flags.BoolVar(&opt.Discovery.Policy.Trust, "discovery-policy-trust", defaultTrust,
		"Enable discovery white-list.")

	opt.Discovery.PollInterval = defaultPollInterval

	opt.Discovery.RediscoverWait = defaultRediscoverWait

	opt.Discovery.Timeout = defaultDiscoveryTimeout

	// options for rate limits
	opt.flags.IntSliceVar(&opt.AccountLevel.Threshold, "account-level", defaultAccountLevel,
		"Rank the accounts then set rate limits for them.")

	opt.flags.Float64Var(&opt.RateLimit.Bandwidth, "ratelimit-bandwidth", defaultBandwidth,
		"Bandwidth of this runtime.")

	opt.flags.Float64Var(&opt.RateLimit.SingleDAGSize, "ratelimit-single-dag-size", defaultSingleDAGSize,
		"Estimated single DAG size to receive from providers.")

	// options for backup
	opt.flags.StringVar(&opt.Backup.EstuaryGateway, "backup-estuary-gateway", defaultEstGateway,
		"Estuary gateway address used to backup metadata files.")

	opt.flags.StringVar(&opt.Backup.ShuttleGateway, "backup-shuttle-gateway", defaultShuttleGateway,
		"Estuary shuttle gateway address used to backup metadata files.")

	_ = opt.viper.BindPFlags(opt.flags)

	return opt
}

// YAML returns yaml string of options, should be called after Parse the options
func (opt *Options) YAML() string {
	return opt.yamlStr
}

// Parse parses all arguments
func (opt *Options) Parse() (string, error) {
	err := opt.flags.Parse(os.Args[1:])
	if err != nil {
		return "", err
	}

	if opt.ShowVersion {
		return version.Long, nil
	}

	if opt.ShowHelp {
		return opt.flags.FlagUsages(), nil
	}

	opt.viper.AutomaticEnv()
	opt.viper.SetEnvPrefix("PD")
	opt.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if opt.ConfigFile != "" {
		opt.viper.SetConfigFile(filepath.Join(opt.PandoRoot, opt.ConfigFile))
		opt.viper.SetConfigType("yaml")
		err := opt.viper.ReadInConfig()
		if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read config file %s failed: %v", opt.ConfigFile, err)
		}
	}

	// NOTE: Workaround because viper does not treat env vars the same as other config.
	// Reference: https://github.com/spf13/viper/issues/188#issuecomment-399518663
	for _, key := range opt.viper.AllKeys() {
		val := opt.viper.Get(key)
		// NOTE: We need to handle map[string]string
		// Reference: https://github.com/spf13/viper/issues/911
		if key == "labels" {
			val = opt.viper.GetStringMapString(key)
		}
		opt.viper.Set(key, val)
	}

	_ = opt.viper.Unmarshal(opt, func(c *mapstructure.DecoderConfig) {
		c.TagName = "yaml"
	})

	err = opt.validate()
	if err != nil {
		return "", err
	}

	buff, err := yaml.Marshal(opt)
	if err != nil {
		return "", fmt.Errorf("marshal config to yaml failed: %v", err)
	}
	opt.yamlStr = string(buff)

	if opt.ShowConfig {
		fmt.Printf("%s", opt.yamlStr)
	}

	return "", nil
}

func (opt *Options) validate() error {
	//ToDo: validate flags
	return nil
}
