package register

import (
	"errors"
	"fmt"
	. "github.com/agiledragon/gomonkey/v2"
	"github.com/kenlabs/pando/test/mock"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/record"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestRegisterRequest(t *testing.T) {
	peerID, privKey, err := mock.GetPrivkyAndPeerID()
	account := "t12345"
	addrs := []string{"/ip4/127.0.0.1/tcp/9999"}
	Convey("test create and load register request", t, func() {
		So(err, ShouldBeNil)
		data, err := MakeRegisterRequest(peerID, privKey, addrs, account)
		So(err, ShouldBeNil)
		peerRec, err := ReadRegisterRequest(data)
		So(err, ShouldBeNil)
		seq0 := peerRec.Seq
		// register again
		data, err = MakeRegisterRequest(peerID, privKey, addrs, account)
		So(err, ShouldBeNil)
		peerRec, err = ReadRegisterRequest(data)
		So(err, ShouldBeNil)
		// seq create from time
		So(seq0, ShouldBeLessThan, peerRec.Seq)

	})
	Convey("test create error and load error", t, func() {
		//Convey("nil addr", func() {
		//	account := "t12345"
		//	peerID, privKey, err := mock.GetPrivkyAndPeerID()
		//	So(err, ShouldBeNil)
		//	data, err := MakeRegisterRequest(peerID, privKey, []string{}, account)
		//	So(err, ShouldResemble, errors.New("missing address"))
		//	So(data, ShouldBeNil)
		//})
		Convey("failed seal the register", func() {
			patch := ApplyFunc(record.Seal, func(rec record.Record, privateKey crypto.PrivKey) (*record.Envelope, error) {
				return nil, errors.New("failed seal")
			})
			defer patch.Reset()
			_, err := MakeRegisterRequest(peerID, privKey, addrs, account)
			So(err, ShouldResemble, fmt.Errorf("could not sign request: failed seal"))
		})
		Convey("failed marshal the register", func() {
			patch := ApplyMethod(reflect.TypeOf(&record.Envelope{}), "Marshal", func(_ *record.Envelope) ([]byte, error) {
				return nil, errors.New("failed")
			})
			defer patch.Reset()
			_, err := MakeRegisterRequest(peerID, privKey, addrs, account)
			So(err, ShouldResemble, fmt.Errorf("could not marshal request register: failed"))
		})
	})
}
