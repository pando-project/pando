package system

import (
	"fmt"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/showwin/speedtest-go/speedtest"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func Test_TestInternetSpeed(t *testing.T) {
	Convey("Test TestInternetSpeed", t, func() {
		downloadSpeed, err := TestInternetSpeed(true)
		if err != nil {
			t.Error(err)
		}
		if downloadSpeed <= 0.5 {
			t.Errorf("download speed is lower than 0.5 Mbps, network environment sucks")
		}
	})
}

func Test_SpeedTestError(t *testing.T) {
	Convey("Test SpeedTest error handle", t, func() {
		patch := gomonkey.ApplyFunc(speedtest.FetchUserInfo, func() (*speedtest.User, error) {
			return nil, fmt.Errorf("unknown error")
		})
		downloadSpeed, err := TestInternetSpeed(true)
		So(downloadSpeed, ShouldEqual, float64(0))
		So(err, ShouldResemble, fmt.Errorf(FailedError, "unknown error"))
		patch.Reset()

		patch = gomonkey.ApplyFunc(speedtest.FetchServerList, func(user *speedtest.User) (speedtest.ServerList, error) {
			return speedtest.ServerList{}, fmt.Errorf("unknown error2")
		})
		downloadSpeed, err = TestInternetSpeed(true)
		So(downloadSpeed, ShouldEqual, float64(0))
		So(err, ShouldResemble, fmt.Errorf(FailedError, "unknown error2"))
		patch.Reset()

		patch = gomonkey.ApplyFunc(serverIsAbroad, func(userCountry string, serverCountry string) bool {
			return true
		})
		downloadSpeed, err = TestInternetSpeed(true)
		patch.Reset()

		patch = gomonkey.ApplyMethod(reflect.TypeOf(&speedtest.Server{}), "DownloadTest", func(_ *speedtest.Server, _ bool) error {
			return fmt.Errorf("unknown error3")
		})
		downloadSpeed, err = TestInternetSpeed(true)
		patch.Reset()

	})
}
