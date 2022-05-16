package log

import (
	"github.com/ipfs/go-log/v2"
	"runtime"
)

func NewSubsystemLogger() *log.ZapEventLogger {
	pc, _, _, _ := runtime.Caller(1)
	callerName := runtime.FuncForPC(pc).Name()

	return log.Logger(callerName)
}
