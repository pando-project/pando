package types

import (
	"bytes"
	"fmt"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"testing"
)

func TestProviderState_MarshalCBOR(t *testing.T) {

	cidlist := make([]cid.Cid, 0)
	for _, hashfun := range []uint64{
		mh.IDENTITY, mh.SHA3, mh.SHA2_256,
	} {
		h1, err := mh.Sum([]byte("TEST"), hashfun, -1)
		if err != nil {
			t.Fatal(err)
		}
		c1 := cid.NewCidV1(cid.Raw, h1)

		h2, err := mh.Sum([]byte("foobar"), hashfun, -1)
		if err != nil {
			t.Fatal(err)
		}
		c2 := cid.NewCidV1(cid.Raw, h2)

		c3, err := c1.Prefix().Sum([]byte("foobar"))
		if err != nil {
			t.Fatal(err)
		}
		if !c2.Equals(c3) {
			t.Fatal("expected CIDs to be equal")
		}
		fmt.Println(c1)
		fmt.Println(c2)
		fmt.Println(c3)
		cidlist = append(cidlist, c1, c2, c3)
	}

	ps := &ProviderState{Cidlist: cidlist}

	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println()

	buf := new(bytes.Buffer)
	if err := ps.MarshalCBOR(buf); err != nil {
		t.Fatal(err)
	}

	psOut := ProviderState{}
	if err := psOut.UnmarshalCBOR(buf); err != nil {
		t.Fatal(err)
	}
	for _, v := range psOut.Cidlist {
		fmt.Println(v.String())
	}

}
