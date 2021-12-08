package command

import (
	"Pando/config"
	"context"
	"encoding/json"

	//"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

//var pando, _ = mock.NewPandoMock()

func TestRegisterMissingInfo(t *testing.T) {
	// Set up a context that is canceled when the command is interrupted
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	os.Setenv(config.EnvDir, tempDir)

	app := &cli.App{
		Name: "indexer",
		Commands: []*cli.Command{
			RegisterCmd,
		},
	}

	err := app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102"})
	if err != nil {
		log.Error(err.Error())
	}
}

func TestRegisterErrorParam(t *testing.T) {
	// Set up a context that is canceled when the command is interrupted
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	os.Setenv(config.EnvDir, tempDir)
	os.Setenv("PANDO", "testurl")

	app := &cli.App{
		Name: "indexer",
		Commands: []*cli.Command{
			RegisterCmd,
		},
	}

	wrongpeeridStr := "wrongpeerid"
	peeridStr := "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"
	privkeyStr := "CAESQPlmGpltNdirG30V5jH79GrPwp2lJz5rUWC/4leCfYMP9mzDYut2ootG0Xx2ZAy7sGVSsEcVMk0e1+qOOf0KHXU="
	wrongpk := "wrongpk"

	err := app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", wrongpeeridStr})
	assert.Contains(t, err.Error(), "could not decode account id")

	err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", wrongpk, "--peerid", peeridStr})
	assert.Contains(t, err.Error(), "could not decode private key")

	err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", "fsdctfvghbjnkctfvgbh"})
	assert.Contains(t, err.Error(), "no such file or directory")

	wrongIdentity := config.Identity{
		PeerID:  wrongpeeridStr,
		PrivKey: privkeyStr,
	}
	_ = struct {
		Name string
		Age  int
	}{Name: "???", Age: 88}

	f, err := os.Create(tempDir + "testconfig")
	assert.NoError(t, err)
	f2, err := os.Create(tempDir + "testconfig2")
	assert.NoError(t, err)
	wrongcfgbytes, err := json.Marshal(wrongIdentity)
	assert.NoError(t, err)
	_, err = f.Write([]byte("abcdefg"))
	assert.NoError(t, err)
	_, err = f2.Write(wrongcfgbytes)
	assert.NoError(t, err)

	err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", tempDir + "testconfig"})
	assert.Contains(t, err.Error(), "invalid character")

	err = app.RunContext(ctx, []string{"pando", "register", "--pa", "/ip4/127.0.0.1/tcp/3102", "--privkey", privkeyStr, "--peerid", "", "-config", tempDir + "testconfig2"})
	assert.Contains(t, err.Error(), "could not decode account id")

}
