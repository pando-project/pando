package config

func DisableTestSpeed() Option {
	return func(config *Config) {
		config.BandWidth = -1
	}
}
