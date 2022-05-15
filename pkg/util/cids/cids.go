package cids

import (
	"fmt"
	"github.com/ipfs/go-cid"
)

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
