//go:build windows

package localapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Microsoft/go-winio"
	"golang.org/x/sys/windows"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func servePipe(t *testing.T, path, sddl string) {
	t.Helper()
	ln, err := listenPipe(path, sddl)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = Serve(ctx, ln, Handler(Funcs{Status: func(context.Context) (proto.LocalStatus, error) { return proto.LocalStatus{MeshIP: "10.77.2.1"}, nil }}))
	}()
	waitPipe(t, path)
}

// waitPipe waits until the pipe takes clients. go-winio's first instance
// takes none until the listener's first Accept makes the next one; until
// then a client finds no pipe at all. Any other answer -- a connection, or
// access denied for a pipe this user may not open -- means it is up.
func waitPipe(t *testing.T, path string) {
	t.Helper()
	wait := 100 * time.Millisecond
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		c, err := winio.DialPipe(path, &wait)
		if err == nil {
			_ = c.Close()
			return
		}
		if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return
		}
	}
	t.Fatalf("%s took no client within 5s", path)
}

func pipeName() string { return fmt.Sprintf(`\\.\pipe\meshctl-test-%d`, time.Now().UnixNano()) }

// The local API's pipe on Windows lets only Administrators and SYSTEM in
// (amendment A7): its SDDL is a protected DACL with exactly those two, and
// the pipe carries it -- from an unelevated prompt the client is refused
// with access denied; from an elevated one it is answered. Either way a
// second agent may not take the pipe, and a name that is not a pipe's is
// refused.
func TestPipeIsForAdministratorsOnly(t *testing.T) {
	sd, err := windows.SecurityDescriptorFromString(PipeSDDL)
	if err != nil {
		t.Fatal(err)
	}
	if got := sd.String(); got != "D:P(A;;GA;;;BA)(A;;GA;;;SY)" {
		t.Fatalf("PipeSDDL reads back as %s, want only Administrators and SYSTEM", got)
	}
	path := pipeName()
	servePipe(t, path, PipeSDDL)
	st, err := NewClient(path).Status(context.Background())
	switch elevated := windows.GetCurrentProcessToken().IsElevated(); {
	case elevated && (err != nil || st.MeshIP != "10.77.2.1"):
		t.Fatalf("an elevated client = %+v, %v", st, err)
	case !elevated && (!errors.Is(err, windows.ERROR_ACCESS_DENIED) || !strings.Contains(err.Error(), "elevated")):
		t.Fatalf("an unelevated client was not refused access: %+v, %v", st, err)
	}
	if _, err := Listen(path); err == nil {
		t.Fatal("a second agent took the pipe")
	}
	if _, err := Listen(`\\.\pipe\a b`); err == nil || !strings.Contains(err.Error(), "named pipe") {
		t.Fatalf("Listen on a bad pipe name = %v", err)
	}
}

// Over a pipe this user may open, the API answers as on a socket.
func TestStatusOverAPipe(t *testing.T) {
	u, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	path := pipeName()
	servePipe(t, path, PipeSDDL+"(A;;GA;;;"+u.User.Sid.String()+")")
	if st, err := NewClient(path).Status(context.Background()); err != nil || st.MeshIP != "10.77.2.1" {
		t.Fatalf("Status over a pipe = %+v, %v", st, err)
	}
	if _, err := listenPipe(path, PipeSDDL); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("a second listener = %v", err)
	}
}
