package types

import (
	"fmt"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/xerrors"
	"io"
)

var lengthBufProviderState = []byte{131}

func (t *ProviderState) MarshalCBOR(w io.Writer) error {
	if t.Cidlist == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write(lengthBufProviderState); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	if len(t.Cidlist) > cbg.MaxLength {
		return xerrors.Errorf("Value cidlist in ProviderState was too long")
	}
	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Cidlist))); err != nil {
		return err
	}
	for _, v := range t.Cidlist {
		if err := cbg.WriteCidBuf(scratch, w, v); err != nil {
			return err
		}
	}
	return nil
}

func (t *ProviderState) UnmarshalCBOR(r io.Reader) error {
	*t = ProviderState{}
	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 3 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	{
		maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
		if err != nil {
			return err
		}

		if extra > cbg.MaxLength {
			return fmt.Errorf("t.cidlist: array too large (%d)", extra)
		}

		if maj != cbg.MajArray {
			return fmt.Errorf("expected cbor array")
		}
		if extra > 0 {
			t.Cidlist = make([]cid.Cid, extra)
		}
		for i := 0; i < int(extra); i++ {

			c, err := cbg.ReadCid(br)
			if err != nil {
				return xerrors.Errorf("failed to read cid field t.cidlist[%d]: %w", i, err)
			}
			t.Cidlist[i] = c
		}
	}
	return nil

}
