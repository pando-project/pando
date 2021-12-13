package config

const (
	defaultShuttleGateway = "https://shuttle-4.estuary.tech"
	defaultEstGateway     = "https://api.estuary.tech"
)

// Backup tracks the configuration of backup in estuary.
type Backup struct {
	EstuaryGateway string
	ShuttleGateway string
	ApiKey         string
}
