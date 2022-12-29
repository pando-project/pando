package system

import (
	"testing"
)

func TestExit(t *testing.T) {
	t.Run("Test Exit status and info", func(t *testing.T) {
		//patch := gomonkey.ApplyFunc(os.Exit, func(code int) {
		//})
		//defer patch.Reset()
		//go func() {
		//	Exit(0, "test error")
		//}()
		//go func() {
		//
		//	Exit(1, "test error2")
		//}()
		//time.Sleep(time.Second)
	})

}
