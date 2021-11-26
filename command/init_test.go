package command

import (
	"Pando/config"
	"context"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Set up a context that is canceled when the command is interrupted
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempDir := t.TempDir()
	os.Setenv(config.EnvDir, tempDir)

	app := &cli.App{
		Name: "pando",
		Commands: []*cli.Command{
			InitCmd,
		},
	}

	badAddr := "ip3/127.0.0.1/tcp/9999"
	err := app.RunContext(ctx, []string{"pando", "init", "-listen-graphsync", badAddr, "-speedtest=false"})
	if err == nil {
		log.Fatal("expected error")
	}

	err = app.RunContext(ctx, []string{"pando", "init", "-listen-graphql", badAddr, "-speedtest=false"})
	if err == nil {
		log.Fatal("expected error")
	}

	goodAddr := "/ip4/127.0.0.1/tcp/7777"
	goodAddr2 := "/ip4/127.0.0.1/tcp/17171"
	args := []string{
		"pando", "init",
		"-listen-graphql", goodAddr,
		"-listen-graphsync", goodAddr2,
		"-speedtest=0",
	}
	err = app.RunContext(ctx, args)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.Load("")
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Addresses.GraphQL != goodAddr {
		t.Error("ingest listen address was not configured")
	}
	if cfg.Addresses.GraphSync != goodAddr2 {
		t.Error("cache size was tno configured")
	}

	t.Log(cfg.Addresses)
}
