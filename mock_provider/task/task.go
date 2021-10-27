package task

import (
	"math/rand"
	"strconv"
)

type FinishedTask struct {
	Miner           string `json:"miner"`
	Status          int    `json:"status"`
	MaxPriceAttoFIL uint64 `json:"max_price_atto_fil"`
	Verified        bool   `json:"verified"`
	FastRetrieval   bool   `json:"fast_retrieval"`
}

func GenMockTask(num int) []*FinishedTask {
	res := make([]*FinishedTask, 0)
	for i := 0; i < num; i++ {
		t := &FinishedTask{
			Miner:           "TestMiner" + strconv.Itoa(rand.Intn(10000)),
			Status:          rand.Intn(5),
			MaxPriceAttoFIL: rand.Uint64(),
			Verified:        i%4 == 0,
			FastRetrieval:   i%5 == 0,
		}
		res = append(res, t)
	}
	return res
}
