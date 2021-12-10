package metadata

import (
	"Pando/config"
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestBackUpFile(t *testing.T) {
	Convey("when send right request then get 200 response", t, func() {
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &config.Backup{
			EstuaryGateway: DefaultEstGateway,
			ShuttleGateway: DefaultShuttleGateway,
		}
		_, err = NewBackupSys(cfg)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 10)

	})
}
