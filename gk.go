package gkgrace

import (
	"os"
	"time"

	"github.com/gogf/gf/os/genv"
)

/*
  types
*/
type Hook func() error

type GraceStatus int

func (that GraceStatus) String() (r string) {
	switch that {
	case 1:
		r = "Exiting"
	case 2:
		r = "Reloading"
	default:
		r = "Unknown"
	}
	return
}

// names of environment variables
const (
	GraceEnvIsChild     = "GRACE_IS_CHILD"      // to mark the child process by "true"
	GraceEnvFdsInSingle = "GRACE_FDS_IN_SINGLE" // single-process mode, add fds to env
)

// offset for extrafiles
const (
	DefaultOffset      = 3
	DefualtMaxWaitTime = 15 * time.Second
)

// Grace status
const (
	GraceUnKnown   GraceStatus = 0
	GraceExiting   GraceStatus = 1
	GraceReloading GraceStatus = 2
)

var IsChildProcess = genv.GetVar(GraceEnvIsChild, false).Bool()

var WorkingDir, _ = os.Getwd()

/*
  interfaces
*/
type IAddress interface {
	GetAddr() *Address
	SetGrace(g *Grace)
}
