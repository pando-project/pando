package command

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

var app = &cli.App{
	Name: "indexer",
	Commands: []*cli.Command{
		IngestCmd,
	},
}

func TestMissingParam(t *testing.T) {
	ctx := context.Background()

	os.Setenv("PANDO", "testurl")

	err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe"})
	assert.Contains(t, err.Error(), "Required flag \"prov\" not set")

	err = app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "-prov"})
	assert.Contains(t, err.Error(), "flag needs an argument: -prov")

	err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe"})
	assert.Contains(t, err.Error(), "Required flag \"prov\" not set")

	err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe", "-prov"})
	assert.Contains(t, err.Error(), "flag needs an argument: -prov")
}

func TestWrongParam(t *testing.T) {
	ctx := context.Background()

	os.Setenv("PANDO", "testurl")

	err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "--prov", "?????/"})
	assert.Contains(t, err.Error(), "failed to parse peer ID")

	err = app.RunContext(ctx, []string{"pando", "ingest", "unsubscribe", "--prov", "?????/"})
	assert.Contains(t, err.Error(), "failed to parse peer ID")
}

func TestNormal(t *testing.T) {
	ctx := context.Background()

	os.Setenv("PANDO", "testurl")
	err := app.RunContext(ctx, []string{"pando", "ingest", "subscribe", "--prov", "12D3KooWSQJeqeks5YAEzAaLdevYNUXYUp7bk9tHt9UXDQkVS3JC"})
	assert.Contains(t, err.Error(), "Get ")
}
