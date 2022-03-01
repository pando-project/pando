package command

import (
	"github.com/kenlabs/pando/pkg/option"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cobra"
	"testing"
	"time"
)

func TestDaemon(t *testing.T) {
	Convey("test start daemon with config", t, func() {
		opt := option.New(nil)
		r, err := opt.Parse()
		if err != nil {
			t.Error(r, err.Error())
		}
		tmp := t.TempDir()
		opt.PandoRoot = tmp
		Opt = opt
		opt.DisableSpeedTest = true
		app := &cobra.Command{}
		app.AddCommand(DaemonCmd())
		app.AddCommand(InitCmd())
		app.SetArgs([]string{"init"})
		_, err = app.ExecuteC()
		if err != nil {
			t.Error(err.Error())
		}
		app.SetArgs([]string{"daemon"})
		errCh := make(chan error)
		go func() {
			_, err = app.ExecuteC()
			errCh <- err
		}()
		time.Sleep(time.Second * 6)
		select {
		case err := <-errCh:
			t.Error(err.Error())
		default:
		}
	})
}
