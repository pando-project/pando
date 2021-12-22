package command

import (
	"Pando/config"
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	Convey("when run init command with wrong parameter then get error", t, func() {
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
		err := app.RunContext(ctx, []string{"pando", "init", "-listen-admin", badAddr, "-speedtest=false"})
		So(err.Error(), ShouldContainSubstring, "bad listen-admin")

		err = app.RunContext(ctx, []string{"pando", "init", "-listen-pando", badAddr, "-speedtest=false"})
		So(err.Error(), ShouldContainSubstring, "bad listen-pando: failed to parse multiaddr")
	})
	Convey("when run init with right parameter then get nil err", t, func() {
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

		goodAddr := "/ip4/127.0.0.1/tcp/7777"
		goodAddr2 := "/ip4/127.0.0.1/tcp/17171"
		args := []string{
			"pando", "init",
			"-listen-pando", goodAddr,
			"-listen-admin", goodAddr2,
			"-speedtest=0",
		}
		err := app.RunContext(ctx, args)
		So(err, ShouldBeNil)
	})
}
