package option

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDuration(t *testing.T) {
	Convey("TestDuration", t, func() {
		Convey("Test Duration helper functions", func() {
			duration := Duration(0)
			durationText := "1s"
			err := duration.UnmarshalText([]byte(durationText))
			if err != nil {
				t.Error(err)
			}
			durationBytes, err := duration.MarshalText()
			if err != nil {
				t.Error(err)
			}
			So(durationBytes, ShouldResemble, []byte(durationText))
			So(duration.String(), ShouldEqual, durationText)
		})
	})
}
