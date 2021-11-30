package registry

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestSequence(t *testing.T) {
	Convey("test sequence", t, func() {
		sq := newSequences(time.Second * 5)
		err := sq.check("dsadsa", uint64(time.Now().Add(-time.Second*6).UnixNano()))
		So(err, ShouldResemble, errors.New("sequence too small"))
		err = sq.check("dsadsa", uint64(time.Now().Add(-time.Second*3).UnixNano()))
		So(err, ShouldBeNil)
		err = sq.check("dsadsa", uint64(time.Now().Add(-time.Second*4).UnixNano()))
		So(err, ShouldResemble, errors.New("sequence less than or equal to last seen"))
	})
}
