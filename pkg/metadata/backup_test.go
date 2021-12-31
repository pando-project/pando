package metadata

import (
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"pando/pkg/option"
	"testing"
	"time"
)

func TestBackUpFile(t *testing.T) {
	Convey("when send right request then get 200 response", t, func() {
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &option.Backup{
			EstuaryGateway: DefaultEstGateway,
			ShuttleGateway: DefaultShuttleGateway,
			ApiKey:         "EST75c4d3bb-d86f-42e4-80da-662d7fbde4c2ARY",
		}
		_, err = NewBackupSys(cfg)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 10)

	})
}
