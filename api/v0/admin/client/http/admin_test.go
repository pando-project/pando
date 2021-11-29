package adminhttpclient

import (
	"Pando/api/v0/admin/model"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateAndRegister(t *testing.T) {
	Convey("test create client and register", t, func() {
		Convey("bad url", func() {
			url := "1`2`2`.`1`3`21/**&*()////foo?query=http://bad"
			_, err := New(url)
			So(err.Error(), ShouldContainSubstring, "invalid character")
		})
		Convey("right url and register", func() {
			c, err := New("http://127.0.0.1")
			So(err, ShouldBeNil)
			gomonkey.ApplyFunc(model.MakeRegisterRequest, func(providerID peer.ID, privateKey crypto.PrivKey, addrs []string, account string) {

			})
		})
	})
}
