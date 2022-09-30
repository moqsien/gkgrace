package gkgrace

import (
	"os"

	"github.com/gogf/gf/os/genv"
)

// names of environment variables
const (
	GraceEnvIsChild     = "GRACE_IS_CHILD"      // to mark the child process by "true"
	GraceEnvFdsInSingle = "GRACE_FDS_IN_SINGLE" // single-process mode, add fds to env
)

// offset for extrafiles
const (
	DefaultOffset = 3
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
