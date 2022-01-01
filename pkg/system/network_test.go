package system

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_TestInternetSpeed(t *testing.T) {
	Convey("Test TestInternetSpeed", t, func() {
		downloadSpeed, err := TestInternetSpeed(true)
		if err != nil {
			t.Error(err)
		}
		if downloadSpeed <= 1 {
			t.Errorf("download speed is lower than 1 Mbps, network environment sucks")
		}
	})
}
