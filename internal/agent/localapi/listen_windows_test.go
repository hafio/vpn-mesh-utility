//go:build windows

package localapi

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

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
}

func pipeName() string { return fmt.Sprintf(`\\.\pipe\meshctl-test-%d`, time.Now().UnixNano()) }

// The local API's pipe on Windows lets only Administrators and SYSTEM in
// (amendment A7): from an unelevated prompt the client is refused; from an
// elevated one it is answered. Either way a second agent may not take the
// pipe, and a name that is not a pipe's is refused.
func TestPipeIsForAdministratorsOnly(t *testing.T) {
	path := pipeName()
	servePipe(t, path, PipeSDDL)
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
	if err != nil {
		t.Fatal(err)
	}
	if dacl := sd.String(); !strings.HasPrefix(dacl, "D:P") || strings.Count(dacl, "(A;") != 2 || !strings.Contains(dacl, ";;;BA)") ||
		!strings.Contains(dacl, ";;;SY)") {
		t.Fatalf("pipe DACL %s, want only Administrators and SYSTEM", dacl)
	}
	st, err := NewClient(path).Status(context.Background())
	switch elevated := windows.GetCurrentProcessToken().IsElevated(); {
	case elevated && (err != nil || st.MeshIP != "10.77.2.1"):
		t.Fatalf("an elevated client = %+v, %v", st, err)
	case !elevated && (err == nil || !strings.Contains(err.Error(), "elevated")):
		t.Fatalf("an unelevated client reached the pipe: %+v, %v", st, err)
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
