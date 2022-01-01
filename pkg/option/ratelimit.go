package option

const (
	defaultBandwidth     = 1.0 // Mbps
	defaultSingleDAGSize = 1.0 // Mb
)

type RateLimit struct {
	Bandwidth     float64 `yaml:"Bandwidth"`
	SingleDAGSize float64 `yaml:"SingleDAGSize"`
}
