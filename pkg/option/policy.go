package option

const (
	defaultAllow = true
	defaultTrust = true
)

type Policy struct {
	Allow       bool     `yaml:"Allow"`
	Except      []string `yaml:"Except"`
	Trust       bool     `yaml:"Trust"`
	TrustExcept []string `yaml:"TrustExcept"`
}
