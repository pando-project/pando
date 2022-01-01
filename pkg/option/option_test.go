package option

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestOptions(t *testing.T) {
	Convey("TestOptions", t, func() {
		opt := New(nil)

		Convey("flag values should be default value before parse function execute", func() {
			So(opt.Discovery.Policy.Allow, ShouldEqual, defaultAllow)
			So(opt.Discovery.LotusGateway, ShouldEqual, defaultLotusGateway)
			So(opt.AccountLevel.Threshold, ShouldResemble, defaultAccountLevel)
			So(opt.RateLimit.SingleDAGSize, ShouldEqual, defaultSingleDAGSize)
		})
	})
}
