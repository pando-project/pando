package handler

import (
	"fmt"
	logging "github.com/ipfs/go-log/v2"
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/option"
)

var logger = logging.Logger("v1ServerHandler")

var (
	InValidRequest = fmt.Errorf("invalid request")
)

type ServerHandler struct {
	Core    *core.Core
	Options *option.Options
}

func New(core *core.Core, opt *option.Options) *ServerHandler {
	return &ServerHandler{
		Core:    core,
		Options: opt,
	}
}
