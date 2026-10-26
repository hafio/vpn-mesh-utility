//go:build windows

package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"golang.org/x/sys/windows/svc"
)

// A test is no Windows service; the service of a mesh is named as
// install.ps1 names it.
func TestNotAService(t *testing.T) {
	if is, err := IsWindowsService(); is || err != nil {
		t.Fatalf("a test is a Windows service: %v, %v", is, err)
	}
	if Name("lab") != "meshctl-agent-lab" {
		t.Fatal(Name("lab"))
	}
}

// executes runs the handler as the service manager would and returns the
// states it reported and its exit code.
func executes(t *testing.T, h *handler, requests ...svc.ChangeRequest) ([]svc.State, uint32) {
	t.Helper()
	r, s, code := make(chan svc.ChangeRequest), make(chan svc.Status, 8), make(chan uint32, 1)
	go func() {
		_, c := h.Execute(nil, r, s)
		code <- c
	}()
	for _, req := range requests {
		r <- req
	}
	c := <-code
	close(s)
	var states []svc.State
	for st := range s {
		states = append(states, st.State)
	}
	return states, c
}

// Stop and Shutdown end the agent's run through its context; Interrogate is
// answered with the current state.
func TestHandlerStopsTheAgent(t *testing.T) {
	for _, cmd := range []svc.Cmd{svc.Stop, svc.Shutdown} {
		stopped := false
		h := &handler{run: func(ctx context.Context) error {
			<-ctx.Done()
			stopped = true
			return nil
		}}
		states, code := executes(t, h, svc.ChangeRequest{Cmd: svc.Interrogate, CurrentStatus: svc.Status{State: svc.Running}}, svc.ChangeRequest{Cmd: cmd})
		want := []svc.State{svc.StartPending, svc.Running, svc.Running, svc.StopPending}
		if code != 0 || h.err != nil || !stopped || !slices.Equal(states, want) {
			t.Fatalf("%v: code %d, err %v, stopped %v, states %v", cmd, code, h.err, stopped, states)
		}
	}
}

// A run that fails ends the service with exit code 1, which the service
// manager's recovery restarts.
func TestHandlerReportsAFailedRun(t *testing.T) {
	h := &handler{run: func(context.Context) error { return errors.New("create wg0: access denied") }}
	states, code := executes(t, h)
	if code != 1 || h.err == nil || !slices.Equal(states, []svc.State{svc.StartPending, svc.Running}) {
		t.Fatalf("code %d, err %v, states %v", code, h.err, states)
	}
}
