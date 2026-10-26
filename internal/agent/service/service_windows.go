//go:build windows

package service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
)

// stopWait is how long the service manager is told a stop may take.
const stopWait = 30 * time.Second

// IsWindowsService reports whether the service manager started this
// process.
func IsWindowsService() (bool, error) { return svc.IsWindowsService() }

// RunWindowsService runs run as the service name until it ends or the
// service is told to stop; Windows' shutdown counts as a stop.
func RunWindowsService(name string, run func(context.Context) error) error {
	h := &handler{run: run}
	if err := svc.Run(name, h); err != nil {
		return fmt.Errorf("run as the %s service: %w", name, err)
	}
	return h.err
}

// handler is the agent as the service manager sees it.
type handler struct {
	run func(context.Context) error
	err error
}

// Execute starts run, reports the service running, and cancels run on
// Stop or Shutdown. A run that fails exits with code 1, so the service
// manager's recovery (install.ps1 sets it) restarts it.
func (h *handler) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- h.run(ctx) }()
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	for {
		select {
		case err := <-done:
			if h.err = err; err != nil {
				return false, 1
			}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending, WaitHint: uint32(stopWait.Milliseconds())}
				cancel()
				h.err = <-done
				return false, 0
			}
		}
	}
}
