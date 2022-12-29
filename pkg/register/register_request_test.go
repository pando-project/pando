package register

import (
	"testing"

	"github.com/pando-project/pando/pkg/api/v1/model"
	"github.com/pando-project/pando/test/mock"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRequest(t *testing.T) {
	peerID, privKey, err := mock.GetPrivkyAndPeerID()
	account := "t12345"
	addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
	t.Run("test create and load register request", func(t *testing.T) {
		asserts := assert.New(t)
		asserts.Nil(err)
		data, err := model.MakeRegisterRequest(peerID, privKey, addrs, account, "provider1")
		asserts.Nil(err)
		peerRec, err := model.ReadRegisterRequest(data)
		asserts.Nil(err)
		seq0 := peerRec.Seq
		// register again
		data, err = model.MakeRegisterRequest(peerID, privKey, addrs, account, "provider2")
		asserts.Nil(err)
		peerRec, err = model.ReadRegisterRequest(data)
		asserts.Nil(err)
		// seq create from time
		asserts.Less(seq0, peerRec.Seq)

	})
	t.Run("test create error and load error", func(t *testing.T) {
		t.Run("failed seal the register", func(t *testing.T) {
			//patch := ApplyFunc(record.Seal, func(rec record.Record, privateKey crypto.PrivKey) (*record.Envelope, error) {
			//	return nil, errors.New("failed seal")
			//})
			//defer patch.Reset()
			//_, err := model.MakeRegisterRequest(peerID, privKey, addrs, account, "provider1")
			//So(err, ShouldResemble, fmt.Errorf("could not sign request: failed seal"))
		})
		t.Run("failed marshal the register", func(t *testing.T) {
			//patch := ApplyMethod(reflect.TypeOf(&record.Envelope{}), "Marshal", func(_ *record.Envelope) ([]byte, error) {
			//	return nil, errors.New("failed")
			//})
			//defer patch.Reset()
			//_, err := model.MakeRegisterRequest(peerID, privKey, addrs, account, "provider2")
			//So(err, ShouldResemble, fmt.Errorf("could not marshal request register: failed"))
		})
	})
}
