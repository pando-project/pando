package metadata

import (
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestBackUpFile(t *testing.T) {
	Convey("when send right request then get 200 response", t, func() {
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		_, err = NewBackupSys("https://api.estuary.tech", "https://shuttle-4.estuary.tech")
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 10)
		//err = bs.backupToEstuary("/Users/zxh/ken-labs/Pando/metadata/tmp/backup.car")
		//So(err, ShouldBeNil)
		//_, err = bs.checkDealForBackup(11726283)
		//So(err, ShouldBeNil)
	})
}
