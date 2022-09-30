package gkgrace

import (
	"net"

	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/v2/container/garray"
	"github.com/moqsien/processes/logger"
)

// Container is listener container
type Container struct {
	Names *garray.SortedStrArray
	Data  *gmap.StrAnyMap
}

func NewContainer() *Container {
	return &Container{
		Names: garray.NewSortedStrArray(true),
	}
}

// Add add listener to container
func (that *Container) Add(name string, l any) {
	if that.Data == nil {
		that.Data = gmap.NewStrAnyMap(true)
	}
	if found := that.Data.Contains(name); found {
		panic("Listener duplicated!")
	}
	switch l.(type) {
	case *net.TCPListener, *net.UnixListener:
		that.AddNull(name)
		that.Data.Set(name, l)
	default:
		logger.Errorf("Listener : %s is not supported!", name)
	}
}

// Add add null listener to container
func (that *Container) AddNull(name string) {
	if that.Names.Search(name) == -1 {
		that.Names.Append(name)
	}
}

// SearchIndex find the index of a listener, return -1 if not found
func (that *Container) SearchIndex(name string) int {
	return that.Names.Search(name)
}
