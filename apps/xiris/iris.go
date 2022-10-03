package xiris

import (
	"crypto/tls"
	"fmt"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/host"
	"github.com/moqsien/gkgrace/apps/base"
	"github.com/moqsien/processes/logger"
)

type IrisGrace struct {
	*iris.Application
	*base.Base
	hostConfigs []host.Configurator
	configs     []iris.Configurator
}

func New() *IrisGrace {
	return &IrisGrace{
		Application: iris.New(),
		Base:        base.New(),
	}
}

type IRVisitor interface {
	ExtraMethod(that *IrisGrace) error
}

func (that *IrisGrace) ExtraMethod(r IRVisitor) {
	if err := r.ExtraMethod(that); err != nil {
		logger.Errorf("'ExtraMethod' errored! err: ", err.Error())
	}
}

func (that *IrisGrace) SetHostConfigs(cnfs ...host.Configurator) {
	that.hostConfigs = append(that.hostConfigs, cnfs...)
}

func (that *IrisGrace) SetConfigs(cnfs ...iris.Configurator) {
	that.configs = append(that.configs, cnfs...)
}

func (that *IrisGrace) Run(certs ...string) error {
	if that.Grace == nil {
		panic("Grace is not set! Please use SetGrace to set it.")
	}
	ln := that.Grace.GetListener(that)
	if ln == nil {
		return fmt.Errorf("Cannot get a listener! ")
	}
	that.SetListener(ln)
	if len(certs) > 1 {
		cert, err := tls.LoadX509KeyPair(certs[0], certs[1])
		if err != nil {
			logger.Errorf("tls: cannot load TLS key pair from certFile=%q and keyFile=%q: %s", certs[0], certs[1], err.Error())
			return err
		}
		config := &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2", "http/1.1"},
		}
		ln = tls.NewListener(ln, config)
	}
	runner := iris.Listener(ln, that.hostConfigs...)
	return that.Application.Run(runner, that.configs...)
}
