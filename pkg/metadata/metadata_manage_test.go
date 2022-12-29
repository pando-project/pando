package metadata_test

import (
	"context"
	"testing"
	"time"

	"github.com/pando-project/pando/pkg/legs"
	. "github.com/pando-project/pando/pkg/metadata"
	"github.com/pando-project/pando/test/mock"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
)

func TestReceiveRecordAndOutUpdate(t *testing.T) {
	t.Run("test metadata manager", func(t *testing.T) {
		pando, err := mock.NewPandoMock()
		asserts := assert.New(t)
		asserts.Nil(err)
		lys := legs.MkLinkSystem(pando.PS, nil, nil)

		t.Run("give records when wait for maxInterval then update and backup", func(t *testing.T) {
			//BackupMaxInterval = time.Second * 3
			pando.Opt.Backup.BackupGenInterval = (time.Second * 3).String()
			mm, err := New(context.Background(), pando.DS, &lys, pando.Registry, &pando.Opt.Backup)
			asserts.Nil(err)
			provider, err := mock.NewMockProvider(pando)
			asserts.Nil(err)
			err = pando.Registry.RegisterOrUpdate(context.Background(), provider.ID, cid.Undef, provider.ID, cid.Undef, false)
			asserts.Nil(err)
			cid1, err := provider.SendMeta(true)
			asserts.Nil(err)
			cid2, err := provider.SendMeta(true)
			asserts.Nil(err)
			cid3, err := provider.SendMeta(true)
			asserts.Nil(err)
			mockRecord := []*MetaRecord{
				{cid1, provider.ID, uint64(time.Now().UnixNano())},
				{cid2, provider.ID, uint64(time.Now().UnixNano())},
				{cid3, provider.ID, uint64(time.Now().UnixNano())},
			}
			recvCh := mm.GetMetaInCh()
			for _, r := range mockRecord {
				recvCh <- r
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			t.Cleanup(func() {
				cancel()
			})

			time.Sleep(time.Second * 5)
			update, _, err := pando.PS.SnapShotStore().GetSnapShotByHeight(ctx, 0)
			providerInfo := pando.Registry.AllProviderInfo()
			asserts.Equal(1, len(providerInfo))
			//asserts.Equal(cid3, providerInfo[0])
			asserts.Nil(err)
			asserts.Equal("", update.PrevSnapShot)
			asserts.Equal(3, len(update.Update[provider.ID.String()].MetaList))
			asserts.Contains(update.Update[provider.ID.String()].MetaList, cid1)
			asserts.Contains(update.Update[provider.ID.String()].MetaList, cid2)
			asserts.Contains(update.Update[provider.ID.String()].MetaList, cid3)
			t.Logf("%#v", update)

			mm.Close()
		})
	})
}
