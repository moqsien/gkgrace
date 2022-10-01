package gkgrace

import (
	"context"
	"os"

	"github.com/moqsien/processes/logger"
)

type sigChan chan struct{}

type deferFunc func(ctx context.Context) sigChan

// ExecuteWithTimeout execute df with timeout duration
func (that *Grace) ExecuteWithTimeout(action string, df deferFunc) {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), that.MaxWaitTime)
	defer cancel()
	select {
	case <-ctxTimeout.Done():
		if err := ctxTimeout.Err(); err != nil {
			logger.Errorf("[process]: %d, [action]: %s, execute hooks timeout! error: %s", os.Getpid(), action, err.Error())
		}
	case <-df(ctxTimeout):
	}
}

/*
  useful hooks for Grace
*/

// SetExitHooksForSingle set hooks called when exiting for single-process mode
func (that *Grace) SetExitHooksForSingle(beforeExit Hook, clearUp ...Hook) {
	that.SingleExitingHook = func() error {
		pid := os.Getpid()
		defer os.Exit(0)
		defer logger.Printf("[Pid]: %d exited.", pid)

		reloadFlag := false
		if that.Status == GraceReloading {
			reloadFlag = true
			logger.Printf("[parent]: %d is exiting...\n", pid)
		} else {
			logger.Printf("[process]: %d is exiting...\n", pid)
		}
		that.Status = GraceExiting

		exitFunc := func(ctx context.Context) sigChan {
			c := make(sigChan)
			go func() {
				defer close(c)
				if !reloadFlag && len(clearUp) > 0 {
					// if this is not reloading, then do sth special for clearups
					if err := clearUp[0](); err != nil {
						logger.Errorf("[Pid]: %d, 'clearUp' execution failed! err: %s", pid, err.Error())
					}
				}
				err := beforeExit()
				if err != nil {
					logger.Errorf("[Pid]: %d, 'beforeExit' execution failed! err: %s", pid, err.Error())
				}
			}()
			return c
		}
		that.ExecuteWithTimeout("exit", exitFunc)
		return nil
	}
}

func (that *Grace) SetExitHooksForMulti(exit Hook) {
	that.MultiExitingHook = exit
}

func (that *Grace) SetReloadHooksForMulti(reload Hook) {
	that.MultiReloadHook = reload
}

func (that *Grace) SetExitHooksForMultiChild(beforeExit Hook, clearUp ...Hook) {
	that.MultiChildExitHook = func() error {
		pid := os.Getpid()
		defer os.Exit(0)
		logger.Printf("[Child process]:%d, is exiting...", pid)
		that.Status = GraceExiting

		exitFunc := func(ctx context.Context) sigChan {
			c := make(sigChan)
			go func() {
				defer close(c)
				// do sth. for clearups
				if len(clearUp) > 0 {
					if err := clearUp[0](); err != nil {
						logger.Errorf("Child[Pid]: %d, 'clearUp' execution failed! err: %s", pid, err.Error())
					}
				}

				err := beforeExit()
				if err != nil {
					logger.Errorf("Child[Pid]: %d, 'beforeExit' execution failed! err: %s", pid, err.Error())
				}
			}()
			return c
		}
		that.ExecuteWithTimeout("exit", exitFunc)
		return nil
	}
}
