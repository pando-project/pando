package metadata

import (
	"github.com/agiledragon/gomonkey/v2"
	logging "github.com/ipfs/go-log/v2"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"os"
	"pando/pkg/option"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func genTmpCarFiles(dir string) error {
	timeStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	f, err := os.OpenFile(path.Join(dir, "/backup-"+timeStr+".car"), os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		return err
	}
	randBytes := make([]byte, 1025)
	rand.Read(randBytes)
	_, err = f.Write(randBytes)
	if err != nil {
		return err
	}
	return nil
}

func TestBackUpFile(t *testing.T) {
	Convey("when send right request then get 200 response", t, func() {
		patch := gomonkey.ApplyGlobalVar(&checkInterval, time.Second*2)
		defer patch.Reset()
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &option.Backup{
			EstuaryGateway: DefaultEstGateway,
			ShuttleGateway: DefaultShuttleGateway,
			APIKey:         "ESTbbbf4557-991b-4f59-89aa-f45eb7b7208aARY",
		}
		tmpDir := t.TempDir()
		patch2 := gomonkey.ApplyGlobalVar(&BackupTmpPath, tmpDir)
		defer patch2.Reset()
		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)

		_, err = NewBackupSys(cfg)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 20)

	})
}

func TestCheckSuccess(t *testing.T) {
	Convey("test back up file successfully", t, func() {
		patch := gomonkey.ApplyGlobalVar(&checkInterval, time.Second*2)
		defer patch.Reset()
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &option.Backup{
			EstuaryGateway: DefaultEstGateway,
			ShuttleGateway: DefaultShuttleGateway,
			APIKey:         "ESTbbbf4557-991b-4f59-89aa-f45eb7b7208aARY",
		}
		tmpDir := t.TempDir()
		patch2 := gomonkey.ApplyGlobalVar(&BackupTmpPath, tmpDir)
		defer patch2.Reset()

		patch3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(&backupSystem{}), "checkDealForBackup", func(_ *backupSystem, _ uint64) (bool, error) {
			return true, nil
		})
		defer patch3.Reset()
		_, err = NewBackupSys(cfg)
		So(err, ShouldBeNil)

		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)
		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)
		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)
		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 10)
	})
}
