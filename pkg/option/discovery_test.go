package option

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDiscoveryFormat(t *testing.T) {
	Convey("test discovery format", t, func() {
		durationText := "1s"
		ds := &Discovery{
			PollInterval:   durationText,
			RediscoverWait: durationText,
			Timeout:        durationText,
		}
		d := ds.PollIntervalInDurationFormat()
		So(d, ShouldResemble, Duration(time.Second))
		d = ds.RediscoverWaitInDurationFormat()
		So(d, ShouldResemble, Duration(time.Second))
		d = ds.TimeoutInDurationFormat()
		So(d, ShouldResemble, Duration(time.Second))
	})
}
