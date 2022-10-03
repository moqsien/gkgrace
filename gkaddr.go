package gkgrace

import "fmt"

// Adress
type Address struct {
	Network string // "tcp", "udp" or "unix"
	Host    string // host ip, "0.0.0.0" by default
	Port    int    // port
	Sock    string // unix domain socket file path if Network is "unix"
}

func (that *Address) String() (s string) {
	if that.Network == "" {
		that.Network = "tcp"
	}
	switch that.Network {
	case "unix", "unixpacket":
		s = fmt.Sprintf("%s@%s", that.Network, that.Sock)
	default:
		if that.Host == "" {
			that.Host = "0.0.0.0"
		}
		s = fmt.Sprintf("%s@%s:%d", that.Network, that.Host, that.Port)
	}
	return
}

func (that *Address) Addr() (s string) {
	if that.Network == "" {
		that.Network = "tcp"
	}
	switch that.Network {
	case "unix", "unixpacket":
		s = that.Sock
	default:
		s = fmt.Sprintf("%s:%d", that.Host, that.Port)
	}
	return
}

func (that *Address) Check() error {
	switch that.Network {
	case "unix", "unixpacket":
		if that.Sock == "" {
			return fmt.Errorf("invalid address!")
		}
	default:
		if that.Port == 0 {
			return fmt.Errorf("invalid address!")
		}
	}
	return nil
}
