package xniogn

import (
	"fmt"

	"github.com/moqsien/gkgrace/apps/base"
	"github.com/moqsien/niogin/httpserver"
	"github.com/moqsien/processes/logger"
)

type NioGrace struct {
	*httpserver.Engine
	*base.Base
}

func New() *NioGrace {
	return &NioGrace{
		Engine: httpserver.New(),
		Base:   base.New(),
	}
}

type INVistitor interface {
	ExtraMethod(that *NioGrace) error
}

func (that *NioGrace) ExtraMethod(n INVistitor) {
	if err := n.ExtraMethod(that); err != nil {
		logger.Errorf("'ExtraMethod' errored! err: ", err.Error())
	}
}

func (that *NioGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.SetListener(ln)
	that.Engine.SetPoll(true)
	if len(certs) > 1 {
		return that.Engine.ServeTLS(ln, certs[0], certs[1])
	}
	return that.Engine.Serve(ln)
}
