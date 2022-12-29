package log

import (
	"runtime"

	"github.com/ipfs/go-log/v2"
)

func NewSubsystemLogger() *log.ZapEventLogger {
	pc, _, _, _ := runtime.Caller(1)
	callerName := runtime.FuncForPC(pc).Name()

	return log.Logger(callerName)
}

func NewSubsystemLoggerWithName(name string) *log.ZapEventLogger {
	return log.Logger(name)
}
