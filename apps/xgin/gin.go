package xgin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moqsien/gkgrace/apps/base"
	"github.com/moqsien/processes/logger"
)

// graceful wrapper for gin
// implementation of IAppBase
type GinGrace struct {
	*gin.Engine
	*base.Base
}

func New() *GinGrace {
	return &GinGrace{
		Engine: gin.New(),
		Base:   base.New(),
	}
}

type IGVisitor interface {
	ExtraMethod(that *GinGrace) error
}

// ExtraMethod visitor pattern, add extra method for GinGrace.
func (that *GinGrace) ExtraMethod(g IGVisitor) {
	if err := g.ExtraMethod(that); err != nil {
		logger.Errorf("'ExtraMethod' errored! err: ", err.Error())
	}
}

func (that *GinGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.SetListener(ln)
	srv := &http.Server{Addr: that.Address.Addr(), Handler: that}
	if len(certs) > 1 {
		// TLS
		return srv.ServeTLS(ln, certs[0], certs[1]) // listener, certFile, keyFile
	}
	// no TLS
	return srv.Serve(ln)
}
