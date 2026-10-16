//go:build windows

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// The Windows service (spec S7.7): meshctl-agent, started with Windows and
// restarted by the service manager when it fails, running `meshctl agent
// run --data <dir>` as LocalSystem.
const (
	// Name is the service's name.
	Name        = "meshctl-agent"
	displayName = "meshctl agent"
	description = "Keeps this machine in its meshctl WireGuard mesh."
	// stopWait bounds waiting for the service to stop.
	stopWait = 30 * time.Second
)

// SCM installs the agent with the Windows service manager. The fields are
// seams for tests; Windows fills them for the host.
type SCM struct {
	// DataDir is the data dir the service runs with.
	DataDir string
	// Executable is the binary the service runs.
	Executable func() (string, error)
}

// Windows is the installer for this host: the service runs the running
// binary unless Executable says otherwise.
func Windows(dataDir string) SCM { return SCM{DataDir: dataDir, Executable: os.Executable} }

// Available reports whether the service can be installed: it takes an
// elevated prompt.
func (s SCM) Available() (bool, string) {
	if !windows.GetCurrentProcessToken().IsElevated() {
		return false, "installing the Windows service needs an elevated prompt (Run as administrator)"
	}
	return true, ""
}

// args is the service's command line after the binary.
func (s SCM) args() []string { return []string{"agent", "run", "--data", s.DataDir} }

// commandLine quotes exe and args the way the service manager reads them.
func commandLine(exe string, args []string) string {
	line := syscall.EscapeArg(exe)
	for _, a := range args {
		line += " " + syscall.EscapeArg(a)
	}
	return line
}

// Install creates the service, or points the existing one at this binary
// and data dir, sets it to restart on failure, and (re)starts it so a
// re-join takes effect.
func (s SCM) Install(ctx context.Context) error {
	exe, err := s.Executable()
	if err != nil {
		return fmt.Errorf("find the meshctl binary: %w", err)
	}
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("open the Windows service manager: %w (run from an elevated prompt)", err)
	}
	defer m.Disconnect()
	h, err := m.OpenService(Name)
	switch {
	case errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST):
		h, err = m.CreateService(Name, exe, mgr.Config{DisplayName: displayName, Description: description, StartType: mgr.StartAutomatic}, s.args()...)
		if err != nil {
			return fmt.Errorf("create the %s service: %w", Name, err)
		}
	case err != nil:
		return fmt.Errorf("open the %s service: %w", Name, err)
	default:
		if err := stop(ctx, h); err != nil {
			h.Close()
			return err
		}
		cfg, err := h.Config()
		if err == nil {
			cfg.BinaryPathName, cfg.DisplayName, cfg.Description, cfg.StartType = commandLine(exe, s.args()), displayName, description, mgr.StartAutomatic
			err = h.UpdateConfig(cfg)
		}
		if err != nil {
			h.Close()
			return fmt.Errorf("update the %s service: %w", Name, err)
		}
	}
	defer h.Close()
	restart := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
		{Type: mgr.ServiceRestart, Delay: time.Minute},
	}
	if err := h.SetRecoveryActions(restart, uint32((24 * time.Hour).Seconds())); err != nil {
		return fmt.Errorf("set the %s service to restart on failure: %w", Name, err)
	}
	// An agent that exits with an error is a failure too, not only a crash.
	if err := h.SetRecoveryActionsOnNonCrashFailures(true); err != nil {
		return fmt.Errorf("set the %s service to restart on failure: %w", Name, err)
	}
	if err := h.Start(); err != nil {
		return fmt.Errorf("start the %s service: %w (`Get-WinEvent -LogName System` says why)", Name, err)
	}
	return nil
}

// Uninstall stops the service and removes it; no service is nothing to do.
func (s SCM) Uninstall(ctx context.Context) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("open the Windows service manager: %w (run from an elevated prompt)", err)
	}
	defer m.Disconnect()
	h, err := m.OpenService(Name)
	switch {
	case errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST):
		return nil
	case err != nil:
		return fmt.Errorf("open the %s service: %w", Name, err)
	}
	defer h.Close()
	if err := stop(ctx, h); err != nil {
		return err
	}
	if err := h.Delete(); err != nil {
		return fmt.Errorf("remove the %s service: %w", Name, err)
	}
	return nil
}

// Stop stops the service when it exists and runs.
func (s SCM) Stop(ctx context.Context) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("open the Windows service manager: %w (run from an elevated prompt)", err)
	}
	defer m.Disconnect()
	h, err := m.OpenService(Name)
	switch {
	case errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST):
		return nil
	case err != nil:
		return fmt.Errorf("open the %s service: %w", Name, err)
	}
	defer h.Close()
	return stop(ctx, h)
}

// stop asks a running service to stop and waits until it has.
func stop(ctx context.Context, h *mgr.Service) error {
	st, err := h.Query()
	if err != nil {
		return fmt.Errorf("query the %s service: %w", Name, err)
	}
	if st.State == svc.Stopped {
		return nil
	}
	if _, err := h.Control(svc.Stop); err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) {
		return fmt.Errorf("stop the %s service: %w", Name, err)
	}
	ctx, cancel := context.WithTimeout(ctx, stopWait)
	defer cancel()
	for {
		if st, err = h.Query(); err != nil {
			return fmt.Errorf("query the %s service: %w", Name, err)
		}
		if st.State == svc.Stopped {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("the %s service did not stop within %s", Name, stopWait)
		case <-time.After(300 * time.Millisecond):
		}
	}
}

// IsWindowsService reports whether the service manager started this
// process.
func IsWindowsService() (bool, error) { return svc.IsWindowsService() }

// RunWindowsService runs run under the service manager until it ends or
// the service is told to stop; Windows' shutdown counts as a stop.
func RunWindowsService(run func(context.Context) error) error {
	h := &handler{run: run}
	if err := svc.Run(Name, h); err != nil {
		return fmt.Errorf("run as the %s service: %w", Name, err)
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
// manager's recovery restarts it.
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
