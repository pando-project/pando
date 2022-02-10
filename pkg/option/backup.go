package option

import "time"

const (
	defaultShuttleGateway    = "https://shuttle-4.estuary.tech"
	defaultEstGateway        = "https://api.estuary.tech"
	defaultBackupGenInterval = time.Minute
	defaultBackupEstInterval = time.Hour * 24
	defaultEstCheckInterval  = time.Hour * 4
)

// Backup tracks the configuration of backup in estuary.
type Backup struct {
	EstuaryGateway    string `yaml:"EstuaryGateway"`
	ShuttleGateway    string `yaml:"ShuttleGateway"`
	APIKey            string `yaml:"APIKey"`
	BackupGenInterval string `yaml:"BackupGenInterval"`
	BackupEstInterval string `yaml:"BackupEstInterval"`
	EstCheckInterval  string `yaml:"EstCheckInterval"`
}
