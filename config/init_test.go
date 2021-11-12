package config

import (
	"bytes"
	"gotest.tools/assert"
	"io"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	cfg, err := Init(io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg.Addresses.P2PAddr, defaultP2PAddr)
	assert.Equal(t, cfg.Addresses.GraphQL, defaultGraphQl)
	assert.Equal(t, cfg.Addresses.MetaData, defaultMetaData)
	assert.Equal(t, cfg.Addresses.GraphSync, defaultGraphSync)
}

func TestSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgFile, err := Filename(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(cfgFile) != tmpDir {
		t.Fatal("wrong root dir", cfgFile)
	}

	cfg, err := Init(io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	cfgBytes, err := Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	err = cfg.Save(cfgFile)
	if err != nil {
		t.Fatal(err)
	}

	cfg2, err := Load(cfgFile)
	if err != nil {
		t.Fatal(err)
	}
	cfg2Bytes, err := Marshal(cfg2)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(cfgBytes, cfg2Bytes) {
		t.Fatal("config data different after being loaded")
	}
}
