package mock

import (
	"math"
)

var (
	Bandwidth     = 100.0
	SingleDAGSize = 2.0
	BaseTokenRate = math.Ceil(0.8 * Bandwidth / SingleDAGSize)
)
