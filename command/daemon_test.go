package command

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/urfave/cli/v2"
	"os"
	"testing"
	"time"
)

var daemonApp = &cli.App{
	Name: "pando",
	Commands: []*cli.Command{
		DaemonCmd,
		InitCmd,
	},
}

func TestDaemonStartAndClose(t *testing.T) {
	tmpdir := t.TempDir()
	err := os.Setenv("PANDO_PATH", tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	Convey("test start daemon and close", t, func() {
		err = daemonApp.RunContext(ctx, []string{"pando", "init", "-speedtest=false"})
		So(err, ShouldBeNil)
		cctx, cncl := context.WithTimeout(ctx, time.Second*3)
		err = daemonApp.RunContext(cctx, []string{"pando", "daemon"})
		So(err, ShouldBeNil)
		cncl()
	})
}
