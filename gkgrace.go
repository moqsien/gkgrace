package gkgrace

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gogf/gf/os/genv"
	"github.com/moqsien/processes/logger"
	"github.com/moqsien/processes/signals"
)

// Grace gracefully restart is supported only when use tcp and unix domain socket
type Grace struct {
	Status             GraceStatus    // status of current process
	Listeners          *Container     // Listeners
	IsChild            bool           // true if in child process
	IsMulti            bool           // true if in multi process mode
	Signal             chan os.Signal // listen for signals
	MaxWaitTime        time.Duration  // maximum wait time
	SingleExitingHook  Hook           // exiting hooks for single-process mode
	MultiChildExitHook Hook           // child process exiting hooks for multi-process mode
	MultiExitingHook   Hook           // master process exiting hooks for multi-process mode
	MultiReloadHook    Hook           // reloading hooks for multi-process mode
}

func New() *Grace {
	return &Grace{
		Status:      GraceUnKnown,
		Listeners:   NewContainer(),
		IsChild:     IsChildProcess,
		Signal:      make(chan os.Signal),
		MaxWaitTime: DefualtMaxWaitTime,
	}
}

func GkListen(addr *Address) (net.Listener, error) {
	var (
		l   net.Listener
		err error
	)

	switch addr.Network {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		l, err = net.Listen(addr.Network, addr.Addr())
		if err != nil {
			logger.Errorf("Listen Errored! err: %s", err.Error())
			return nil, err
		}
	default:
		return nil, fmt.Errorf("Network: %s is not supported!", addr.Network)
	}
	return l, nil
}

// SetToMulti enable multi-process mode
func (that *Grace) SetToMulti() {
	that.IsMulti = true
}

// SetMaxWait set max wait time
func (that *Grace) SetMaxWait(t time.Duration) {
	that.MaxWaitTime = t
}

// Register register a listener before running
func (that *Grace) Register(a IAddress) error {
	addr := a.GetAddr()
	if addr.Host == "" && addr.Sock != "" {
		addr.Host = "0.0.0.0"
	}
	// listener is initialized only in master process for multi-process mode
	if !that.IsChild && that.IsMulti {
		l, err := GkListen(addr)
		if l != nil {
			that.Listeners.Add(addr.String(), l)
			a.SetGrace(that)
		}
		return err
	} else {
		switch addr.Network {
		case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
			that.Listeners.AddNull(addr.String())
			a.SetGrace(that)
			return nil
		default:
			return fmt.Errorf("Network: %s is not supported!", addr.Network)
		}
	}
}

func (that *Grace) GetListener(a IAddress) (l net.Listener) {
	addr := a.GetAddr()
	if that.IsMulti {

	} else {
		// single-process mode
		if !that.IsChild {
			// master
			l, _ = GkListen(addr)
		} else {
			// child
			if offset := that.GetOffsetFromEnv(addr); offset != -1 {
				logger.Printf("[offset]: %d, file inherited from [parent]: %d", offset, os.Getppid())
				l, _ = net.FileListener(os.NewFile(uintptr(offset), addr.String()))
			} else {
				l, _ = GkListen(addr)
			}
		}
	}
	if l != nil {
		that.Listeners.Add(addr.String(), l) // register listener
	}
	return
}

// GetExtrafiles get extrafiles that child process will inherite from
func (that *Grace) GetExtrafiles() (result []*os.File) {
	that.Listeners.Names.Iterator(func(_ int, v string) bool {
		l := that.Listeners.Data.Get(v)
		switch l.(type) {
		case *net.TCPListener:
			file, _ := l.(*net.TCPListener).File()
			result = append(result, file)
		case *net.UnixListener:
			file, _ := l.(*net.UnixListener).File()
			result = append(result, file)
		default:
			fmt.Printf("Listener : %v is not supported!", l)
		}
		return true
	})
	return
}

// GenerateOffsets generate extrafiles offsets map for single-process mode
func (that *Grace) GenerateOffsets() map[string]string {
	offsets := make(map[string]string)
	that.Listeners.Names.Iterator(func(k int, v string) bool {
		key := fmt.Sprintf("%x", md5.Sum([]byte(v)))
		offsets[key] = strconv.Itoa(k + DefaultOffset)
		return true
	})
	return offsets
}

// GetOffsetFromEnv read extrafile offset from environment for single-process mode
func (that *Grace) GetOffsetFromEnv(addr *Address) (offset int) {
	envName := fmt.Sprintf("%x", md5.Sum([]byte(addr.String())))
	return genv.GetVar(envName, -1).Int()
}

// ReloadSingle reload process for single-process mode
func (that *Grace) ReloadSingle() {
	if !that.IsMulti {
		ex, _ := os.Executable()
		cmd := exec.Command(ex)
		cmd.Args = []string{ex}
		cmd.Args = append(cmd.Args, os.Args[1:]...)
		cmd.ExtraFiles = that.GetExtrafiles()
		genv.SetMap(map[string]string{GraceEnvIsChild: "true"}) // to mark the child process by "true"
		genv.SetMap(that.GenerateOffsets())
		cmd.Env = genv.All()
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = WorkingDir
		// cmd.SysProcAttr = &syscall.SysProcAttr{Foreground: true, Noctty: false}
		if err := cmd.Start(); err != nil {
			logger.Errorf("Restart process failed: %s", err.Error())
		}
	}
}

// NotifyParent notify parent process to exit in child
func (that *Grace) NotifyParent() {
	if IsChildProcess {
		parentPid := syscall.Getppid()
		if parentPid != 1 {
			if err := signals.KillPid(parentPid, signals.ToSignal("SIGTERM"), false); err != nil {
				logger.Errorf("failed to send signal to parent process, error: %s", err.Error())
				return
			}
			logger.Printf("Gracefully restarting, child[%d] sent 'SIGTERM' to parent[%d]", syscall.Getpid(), parentPid)
		}
	}
}

func (that *Grace) WaitForSingle() {
	signal.Notify(
		that.Signal,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
	)
	for {
		sig := <-that.Signal
		if that.SingleExitingHook == nil {
			that.SingleExitingHook = func() error {
				logger.Printf("[Parent]: %d, is exiting...", os.Getpid())
				os.Exit(0)
				return nil
			}
		}
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGABRT:
			signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
			that.Status = GraceExiting
			that.MaxWaitTime = time.Second // force to exit within 1 second.
			that.SingleExitingHook()
			continue
		case syscall.SIGQUIT:
			signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
			that.Status = GraceExiting
			that.SingleExitingHook()
			continue
		case syscall.SIGUSR2:
			that.ReloadSingle()
			that.Status = GraceReloading
			continue
		default:
		}
	}
}

func (that *Grace) WaitForMulti() {
	pid := os.Getpid()
	if that.IsChild {
		signal.Notify(
			that.Signal,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGKILL,
			syscall.SIGTERM,
			syscall.SIGABRT,
		)
		for {
			sig := <-that.Signal
			logger.Printf("[Child process]: %d, recieve [signal]: %s", pid, sig.String())
			if that.MultiChildExitHook == nil {
				that.MultiChildExitHook = func() error {
					logger.Printf("Child[%d] is exiting...", os.Getpid())
					os.Exit(0)
					return nil
				}
			}
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGABRT:
				signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
				that.MaxWaitTime = time.Second // force to exit within 1 second.
				that.MultiChildExitHook()
				continue
			case syscall.SIGQUIT:
				signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
				that.MultiChildExitHook()
				continue
			default:
			}
		}
	} else {
		signal.Notify(
			that.Signal,
			syscall.SIGINT,
			syscall.SIGQUIT,
			syscall.SIGKILL,
			syscall.SIGTERM,
			syscall.SIGABRT,
			syscall.SIGUSR1,
			syscall.SIGUSR2,
		)
		for {
			sig := <-that.Signal
			logger.Printf("[Master process]: %d recieve [signal]: %s", pid, sig.String())
			switch sig {
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM:
				signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
				that.MaxWaitTime = time.Second // force to exit within 1 second.
				if that.MultiExitingHook == nil {
					logger.Errorf("'MultiExitingHook' is not set!")
					continue
				}
				that.MultiExitingHook()
				continue
			case syscall.SIGQUIT:
				signal.Reset(syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
				if that.MultiExitingHook == nil {
					logger.Errorf("'MultiExitingHook' is not set!")
					continue
				}
				that.MultiExitingHook()
				continue
			case syscall.SIGUSR2:
				that.MultiReloadHook()
				continue
			default:
			}
		}
	}
}

// Wait wait for signal to come
func (that *Grace) Wait() {
	if that.IsMulti {
		that.WaitForMulti()
	} else {
		that.NotifyParent()
		that.WaitForSingle()
	}
}
