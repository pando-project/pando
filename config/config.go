package config

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

type Config struct {
	Addresses Addresses
}

const (
	DefaultPathName   = ".pando"
	DefaultPathRoot   = "~/" + DefaultPathName
	DefaultConfigFile = "config"
)

var (
	ErrNotInitialized = errors.New("not initialized")
)

// Load reads the json-serialized config at the specified path
func Load(filePath string) (*Config, error) {
	var err error
	if filePath == "" {
		configRoot, err := homedir.Expand(DefaultPathRoot)
		if err != nil {
			return nil, err
		}
		filePath = filepath.Join(configRoot, DefaultConfigFile)
	}

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotInitialized
		}
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, err
}
