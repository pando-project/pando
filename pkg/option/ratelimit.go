package option

const (
	defaultEnable        = false
	defaultBandwidth     = 1.0 // Mbps
	defaultSingleDAGSize = 1.0 // Mb
)

type RateLimit struct {
	Enable        bool    `yaml:"Enable"`
	Bandwidth     float64 `yaml:"Bandwidth"`
	SingleDAGSize float64 `yaml:"SingleDAGSize"`
}
