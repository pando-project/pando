package cids

import (
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/kenlabs/pando-store/pkg/types/store"
	"time"
)

func DecodeAndPadSnapShotList(cidStrList []string) (*store.SnapShotList, error) {
	if len(cidStrList) == 0 {
		return nil, fmt.Errorf("cidStrList cannot be empty")
	}

	res := &store.SnapShotList{}
	res.Length = len(cidStrList)

	for _, cidStr := range cidStrList {
		cidDecoded, err := cid.Decode(cidStr)
		if err != nil {
			return nil, fmt.Errorf("decode cidStr %s failed, err: %v", cidStr, err)
		}
		res.List = append(res.List, struct {
			CreatedTime uint64
			SnapShotCid cid.Cid
		}{CreatedTime: uint64(time.Now().UnixNano()), SnapShotCid: cidDecoded})
	}

	return res, nil
}

func DecodeCidStrList(cidStrList []string) ([]cid.Cid, error) {
	if len(cidStrList) == 0 {
		return nil, fmt.Errorf("cidStrList cannot be empty")
	}

	var cidList []cid.Cid
	for _, cidStr := range cidStrList {
		cidDecoded, err := cid.Decode(cidStr)
		if err != nil {
			return nil, fmt.Errorf("decode cidStr %s failed, err: %v", cidStr, err)
		}
		cidList = append(cidList, cidDecoded)
	}

	return cidList, nil
}
