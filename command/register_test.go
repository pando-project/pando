package command

import (
	"Pando/config"
	"Pando/test/mock"
	"context"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

var pando, _ = mock.NewPandoMock()

func TestRegister(t *testing.T) {
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

	err := app.RunContext(ctx, []string{"pando", "register", "--pa", "???? 123.31231"})
	if err != nil {
		log.Error(err.Error())
	}
}
