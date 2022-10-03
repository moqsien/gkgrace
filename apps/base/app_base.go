package base

import (
	"net"

	"github.com/gogf/gf/os/gtime"
	"github.com/moqsien/gkgrace"
)

type IAppBase interface {
	SetAddr(a *gkgrace.Address)
	GetAddr() *gkgrace.Address
	SetGrace(g *gkgrace.Grace)
	SetListener(l net.Listener)
	FetchListener() net.Listener
	Run(s ...string) error
}

type IApp interface {
	IAppBase
	Name() string
	Execute() error
	Exit() error
}

type AppContainer struct {
	App       IApp
	StartTime *gtime.Time // app start time
	StopTime  *gtime.Time // app stop time
	State     int         // status
}

type Base struct {
	Grace    *gkgrace.Grace
	Address  *gkgrace.Address
	listener net.Listener
}

func New() *Base {
	return &Base{}
}

func (that *Base) FetchListener() net.Listener {
	return that.listener
}

func (that *Base) SetListener(l net.Listener) {
	that.listener = l
}

func (that *Base) SetAddr(addr *gkgrace.Address) {
	if addr.Network == "" {
		addr.Network = "tcp"
	}
	that.Address = addr
}

func (that *Base) GetAddr() *gkgrace.Address {
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

func (that *Base) SetGrace(grace *gkgrace.Grace) {
	that.Grace = grace
}
