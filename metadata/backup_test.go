package metadata

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestBackUpFile(t *testing.T) {
	Convey("when send right request then get 200 response", t, func() {
		bs, err := NewBackupSys("https://shuttle-4.estuary.tech", "Bearer EST75c4d3bb-d86f-42e4-80da-662d7fbde4c2ARY")
		So(err, ShouldBeNil)
		err = bs.backupToEstuary_("/Users/zxh/ken-labs/Pando/metadata/tmp/backup.car")
		So(err, ShouldBeNil)
	})
}
