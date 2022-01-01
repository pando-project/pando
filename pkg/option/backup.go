package option

const (
	defaultShuttleGateway = "https://shuttle-4.estuary.tech"
	defaultEstGateway     = "https://api.estuary.tech"
)

// Backup tracks the configuration of backup in estuary.
type Backup struct {
	EstuaryGateway string `yaml:"EstuaryGateway"`
	ShuttleGateway string `yaml:"ShuttleGateway"`
	ApiKey         string `yaml:"ApiKey"`
}
