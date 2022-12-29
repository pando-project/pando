package metadata_test

import (
	"context"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/pando-project/pando/pkg/legs"
	"github.com/pando-project/pando/test/mock"

	golegs "github.com/filecoin-project/go-legs"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/stretchr/testify/assert"
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
	t.Run("when send right request then get 200 response", func(t *testing.T) {
		//ToDo: mock metadata
		//patch := gomonkey.ApplyGlobalVar(&metadata.CheckInterval, time.Second*2)
		//defer patch.Reset()
		//cfg := &option.Backup{
		//	EstuaryGateway:    metadata.DefaultEstGateway,
		//	ShuttleGateway:    metadata.DefaultShuttleGateway,
		//	BackupGenInterval: time.Second.String(),
		//	BackupEstInterval: time.Second.String(),
		//	EstCheckInterval:  (time.Hour * 24).String(),
		//	APIKey:            apiKey,
		//}
		//tmpDir := t.TempDir()
		//patch2 := gomonkey.ApplyGlobalVar(&metadata.BackupTmpPath, tmpDir)
		//defer patch2.Reset()
		//err := genTmpCarFiles(tmpDir)
		//asserts.Nil(err)
		//
		//_, err = metadata.NewBackupSys(cfg)
		//asserts.Nil(err)
		//time.Sleep(time.Second * 20)

	})
}

func TestCheckSuccess(t *testing.T) {
	t.Run("test back up file successfully", func(t *testing.T) {
		//patch := gomonkey.ApplyGlobalVar(&metadata.CheckInterval, time.Second*2)
		//defer patch.Reset()
		//cfg := &option.Backup{
		//	EstuaryGateway:    metadata.DefaultEstGateway,
		//	ShuttleGateway:    metadata.DefaultShuttleGateway,
		//	BackupGenInterval: time.Second.String(),
		//	BackupEstInterval: time.Second.String(),
		//	EstCheckInterval:  (time.Hour * 24).String(),
		//	APIKey:            apiKey,
		//}
		//tmpDir := t.TempDir()
		//patch2 := patch.ApplyGlobalVar(&metadata.BackupTmpPath, tmpDir)
		//defer patch2.Reset()
		//
		//patch3 := patch.ApplyPrivateMethod(reflect.TypeOf(&metadata.BackupSystem{}), "checkDealForBackup", func(_ *metadata.BackupSystem, _ uint64) (bool, error) {
		//	return true, nil
		//})
		//defer patch3.Reset()
		//_, err := metadata.NewBackupSys(cfg)
		//asserts.Nil(err)
		//
		//err = genTmpCarFiles(tmpDir)
		//asserts.Nil(err)
		//err = genTmpCarFiles(tmpDir)
		//asserts.Nil(err)
		//err = genTmpCarFiles(tmpDir)
		//asserts.Nil(err)
		//err = genTmpCarFiles(tmpDir)
		//asserts.Nil(err)
		//time.Sleep(time.Second * 10)
	})
}

func TestBackUpWithStopLink(t *testing.T) {
	t.Run("when set selector and stop link then get right car file", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)
		provider, err := mock.NewMockProvider(pando)
		asserts.Nil(err)
		var cids []cid.Cid
		for i := 0; i < 5; i++ {
			c, err := provider.SendMeta(true)
			cids = append(cids, c)
			asserts.Nil(err)
			t.Logf("send meta[cid:%s]", c.String())
		}
		time.Sleep(time.Second * 3)

		for i := 0; i < 5; i++ {
			_, err = pando.PS.Get(context.Background(), cids[i])
			asserts.Nil(err)
		}

		linksys := legs.MkLinkSystem(pando.PS, nil, nil)
		ss := golegs.ExploreRecursiveWithStopNode(selector.RecursionLimit{}, nil, cidlink.Link{cids[4]})
		tmpdir := t.TempDir()
		f, err := os.OpenFile(path.Join(tmpdir, "test1.car"), os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		_, err = car.TraverseV1(context.Background(), &linksys, cids[4], ss, f)
		asserts.Nil(err)

	})
}
