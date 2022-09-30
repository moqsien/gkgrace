package xgin

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moqsien/gkgrace"
)

// graceful wrapper for gin
// implementation of IAppBase
type GinGrace struct {
	*gin.Engine
	Grace    *gkgrace.Grace
	Address  *gkgrace.Address
	Listener net.Listener
}

func New() *GinGrace {
	return &GinGrace{
		Engine: gin.New(),
	}
}

func (that *GinGrace) SetAddr(addr *gkgrace.Address) {
	that.Address = addr
}

func (that *GinGrace) GetAddr() *gkgrace.Address {
	if that.Address == nil {
		// default addr
		that.Address = &gkgrace.Address{
			Network: "tcp",
			Host:    "0.0.0.0",
			Port:    8080,
		}
	}
	return that.Address
}

func (that *GinGrace) SetGrace(grace *gkgrace.Grace) {
	that.Grace = grace
}

func (that *GinGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.Listener = ln
	srv := &http.Server{Addr: that.Address.Addr(), Handler: that}
	if len(certs) > 1 {
		// TLS
		return srv.ServeTLS(ln, certs[0], certs[1]) // listener, certFile, keyFile
	}
	// no TLS
	return srv.Serve(ln)
}