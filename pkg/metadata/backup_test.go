package metadata_test

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-car/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/kenlabs/pando/pkg/legs"
	"github.com/kenlabs/pando/pkg/metadata"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/test/mock"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"
)

var apiKey = os.Getenv("APIKEY")

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
		patch := gomonkey.ApplyGlobalVar(&metadata.CheckInterval, time.Second*2)
		defer patch.Reset()
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &option.Backup{
			EstuaryGateway:    metadata.DefaultEstGateway,
			ShuttleGateway:    metadata.DefaultShuttleGateway,
			BackupGenInterval: time.Second.String(),
			BackupEstInterval: time.Second.String(),
			EstCheckInterval:  (time.Hour * 24).String(),
			APIKey:            apiKey,
		}
		tmpDir := t.TempDir()
		patch2 := gomonkey.ApplyGlobalVar(&metadata.BackupTmpPath, tmpDir)
		defer patch2.Reset()
		err = genTmpCarFiles(tmpDir)
		So(err, ShouldBeNil)

		_, err = metadata.NewBackupSys(cfg)
		So(err, ShouldBeNil)
		time.Sleep(time.Second * 20)

	})
}

func TestCheckSuccess(t *testing.T) {
	Convey("test back up file successfully", t, func() {
		patch := gomonkey.ApplyGlobalVar(&metadata.CheckInterval, time.Second*2)
		defer patch.Reset()
		err := logging.SetLogLevel("meta-manager", "debug")
		So(err, ShouldBeNil)
		cfg := &option.Backup{
			EstuaryGateway:    metadata.DefaultEstGateway,
			ShuttleGateway:    metadata.DefaultShuttleGateway,
			BackupGenInterval: time.Second.String(),
			BackupEstInterval: time.Second.String(),
			EstCheckInterval:  (time.Hour * 24).String(),
			APIKey:            apiKey,
		}
		tmpDir := t.TempDir()
		patch2 := patch.ApplyGlobalVar(&metadata.BackupTmpPath, tmpDir)
		defer patch2.Reset()

		patch3 := patch.ApplyPrivateMethod(reflect.TypeOf(&metadata.BackupSystem{}), "checkDealForBackup", func(_ *metadata.BackupSystem, _ uint64) (bool, error) {
			return true, nil
		})
		defer patch3.Reset()
		_, err = metadata.NewBackupSys(cfg)
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

func TestBackUpWithStopLink(t *testing.T) {
	Convey("when set selector and stop link then get right car file", t, func() {
		pando, err := mock.NewPandoMock()
		So(err, ShouldBeNil)
		provider, err := mock.NewMockProvider(pando)
		So(err, ShouldBeNil)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(true)
			cids = append(cids, c)
			So(err, ShouldBeNil)
			t.Logf("send meta[cid:%s]", c.String())
		}
		time.Sleep(time.Second * 3)

		for i := 0; i < 5; i++ {
			_, err = pando.BS.Get(context.Background(), cids[i])
			So(err, ShouldBeNil)
		}

		linksys := legs.MkLinkSystem(pando.BS, nil, nil)
		ss := golegs.ExploreRecursiveWithStopNode(selector.RecursionLimit{}, nil, cidlink.Link{cids[4]})
		f, err := os.OpenFile("./test1.car", os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		_, err = car.TraverseV1(context.Background(), &linksys, cids[4], ss, f)
		So(err, ShouldBeNil)

	})
}
