package system

import (
	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

func TestExit(t *testing.T) {
	convey.Convey("Test Exit status and info", t, func() {
		patch := gomonkey.ApplyFunc(os.Exit, func(code int) {
		})
		defer patch.Reset()
		go func() {
			Exit(0, "test error")
		}()
		go func() {

			Exit(1, "test error2")
		}()
		time.Sleep(time.Second)
	})

}
