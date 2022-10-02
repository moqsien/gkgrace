package xfiber

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/moqsien/gkgrace"
	"github.com/moqsien/processes/logger"
)

type FiberGrace struct {
	*fiber.App
	Grace    *gkgrace.Grace
	Address  *gkgrace.Address
	listener net.Listener
}

func New() *FiberGrace {
	return &FiberGrace{
		App: fiber.New(),
	}
}

type IFVistor interface {
	ExtraMethod(that *FiberGrace) error
}

func (that *FiberGrace) ExtraMethod(f IFVistor) {
	if err := f.ExtraMethod(that); err != nil {
		logger.Errorf("'ExtraMethod' errored! err: ", err.Error())
	}
}

func (that *FiberGrace) Listener() net.Listener {
	return that.listener
}

func (that *FiberGrace) SetAddr(addr *gkgrace.Address) {
	that.Address = addr
}

func (that *FiberGrace) GetAddr() *gkgrace.Address {
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

func (that *FiberGrace) SetGrace(grace *gkgrace.Grace) {
	that.Grace = grace
}

func (that *FiberGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.listener = ln
	if len(certs) > 1 {
		cert, err := tls.LoadX509KeyPair(certs[0], certs[1])
		if err != nil {
			logger.Errorf("tls: cannot load TLS key pair from certFile=%q and keyFile=%q: %s", certs[0], certs[1], err.Error())
			return err
		}
		config := &tls.Config{
			MinVersion:     tls.VersionTLS12,
			Certificates:   []tls.Certificate{cert},
			GetCertificate: (&fiber.TLSHandler{}).GetClientInfo,
		}
		ln = tls.NewListener(ln, config)
	}
	return that.App.Listener(ln)
}
