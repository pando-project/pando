package controller

import (
	"github.com/kenlabs/pando/pkg/api/core"
	"github.com/kenlabs/pando/pkg/option"
	"github.com/kenlabs/pando/pkg/util/log"
)

var logger = log.NewSubsystemLogger()

type Controller struct {
	Core    *core.Core
	Options *option.DaemonOptions
}

func New(core *core.Core, opt *option.DaemonOptions) *Controller {
	return &Controller{
		Core:    core,
		Options: opt,
	}
}
