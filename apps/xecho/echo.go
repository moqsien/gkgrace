package xecho

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	"github.com/moqsien/gkgrace/apps/base"
	"github.com/moqsien/processes/logger"
)

const (
	Version = echo.Version
	website = "https://echo.labstack.com"
	banner  = `
	____    __
   / __/___/ /  ___
  / _// __/ _ \/ _ \
 /___/\__/_//_/\___/ %s
 High performance, minimalist Go web framework
 %s
 ____________________________________O/_______
									 O\
 `
)

type EchoGrace struct {
	*echo.Echo
	*base.Base
	colorer      *color.Color
	startupMutex sync.RWMutex
}

func New() *EchoGrace {
	return &EchoGrace{
		Echo:    echo.New(),
		Base:    base.New(),
		colorer: color.New(),
	}
}

type IEVisitor interface {
	ExtraMethod(that *EchoGrace) error
}

func (that *EchoGrace) ExtraMethod(e IEVisitor) {
	if err := e.ExtraMethod(that); err != nil {
		logger.Errorf("'ExtraMethod' errored! err: ", err.Error())
	}
}

func (that *EchoGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.SetListener(ln)

	that.Echo.HideBanner = true
	that.Echo.Server.Addr = that.GetAddr().Addr()

	if len(certs) == 0 {
		that.startupMutex.Lock()
		if err := that.configServer(that.Echo.Server); err != nil {
			that.startupMutex.Unlock()
			return err
		}
		that.startupMutex.Unlock()
		return that.Echo.Server.Serve(that.Echo.Listener)
	} else if len(certs) > 1 {
		that.startupMutex.Lock()
		s := that.Echo.TLSServer
		s.TLSConfig = new(tls.Config)
		s.TLSConfig.Certificates = make([]tls.Certificate, 1)
		cert, err := tls.LoadX509KeyPair(certs[0], certs[1])
		if err != nil {
			that.startupMutex.Unlock()
			return err
		}
		s.TLSConfig.Certificates[0] = cert
		that.configTLS()
		if err := that.configServer(s); err != nil {
			that.startupMutex.Unlock()
			return err
		}
		that.startupMutex.Unlock()
		return s.Serve(that.Echo.TLSListener)
	}
	return nil
}

func (that *EchoGrace) configServer(s *http.Server) error {
	// Setup
	that.colorer.SetOutput(that.Echo.Logger.Output())
	s.ErrorLog = that.Echo.StdLogger
	s.Handler = that.Echo
	if that.Echo.Debug {
		that.Echo.Logger.SetLevel(log.DEBUG)
	}

	if !that.Echo.HideBanner {
		that.colorer.Printf(banner, that.colorer.Red("v"+Version), that.colorer.Blue(website))
	}

	if s.TLSConfig == nil {
		if that.Echo.Listener == nil {
			that.Echo.Listener = that.FetchListener()
		}
		if !that.Echo.HidePort {
			that.colorer.Printf("⇨ http server started on %s\n", that.colorer.Green(that.Echo.Listener.Addr()))
		}
		return nil
	}

	if that.Echo.TLSListener == nil {
		that.Echo.TLSListener = tls.NewListener(that.FetchListener(), s.TLSConfig)
	}
	if !that.Echo.HidePort {
		that.colorer.Printf("⇨ https server started on %s\n", that.colorer.Green(that.TLSListener.Addr()))
	}
	return nil
}

func (that *EchoGrace) configTLS() {
	s := that.Echo.TLSServer
	s.Addr = that.GetAddr().Addr()
	if !that.Echo.DisableHTTP2 {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "h2")
	}
}
