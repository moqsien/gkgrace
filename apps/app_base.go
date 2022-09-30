package apps

import (
	"github.com/gogf/gf/os/gtime"
	"github.com/moqsien/gkgrace"
)

type IAppBase interface {
	SetAddr(a *gkgrace.Address)
	GetAddr() *gkgrace.Address
	SetGrace(g *gkgrace.Grace)
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
