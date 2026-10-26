package state

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// host.json, session.json and refused.json (amendment A50) read back as
// written, none is not an error, and a damaged one is reported.
func TestSessionFiles(t *testing.T) {
	d := open(t)
	if _, ok, err := d.Host(); ok || err != nil {
		t.Fatalf("no host.json = %v, %v", ok, err)
	}
	if _, ok, err := d.Session(); ok || err != nil {
		t.Fatalf("no session.json = %v, %v", ok, err)
	}
	if _, ok, err := d.Refused(); ok || err != nil {
		t.Fatalf("no refused.json = %v, %v", ok, err)
	}
	h := Host{Machine: "m"}
	s := Session{Session: testfix.Session(1), Ticket: testfix.Ticket(1), Heartbeat: 30, Released: true}
	r := Refused{Code: "foreign", At: 7, Message: "another machine", Token: "fp"}
	for _, err := range []error{d.SaveHost(h), d.SaveSession(s), d.SaveRefused(r)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got, ok, err := d.Host(); !ok || err != nil || got != h {
		t.Fatalf("host.json = %+v, %v, %v", got, ok, err)
	}
	if got, ok, err := d.Session(); !ok || err != nil || got != s {
		t.Fatalf("session.json = %+v, %v, %v", got, ok, err)
	}
	if got, ok, err := d.Refused(); !ok || err != nil || got != r {
		t.Fatalf("refused.json = %+v, %v, %v", got, ok, err)
	}
	for _, err := range []error{d.ClearSession(), d.ClearRefused(), d.ClearSession()} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, ok, _ := d.Session(); ok {
		t.Fatal("session.json survived ClearSession")
	}
	// A session of the wrong shape, and files that are not JSON.
	if err := d.SaveSession(Session{Session: "short", Ticket: testfix.Ticket(1)}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := d.Session(); err == nil || !strings.Contains(err.Error(), "says hello again") {
		t.Fatalf("a malformed session = %v", err)
	}
	for _, name := range []string{HostFile, RefusedFile} {
		if err := os.WriteFile(d.file(name), []byte("{"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := d.Host(); err == nil || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("a damaged host.json = %v", err)
	}
	if _, _, err := d.Refused(); err == nil {
		t.Fatal("a damaged refused.json was read")
	}
}

// Wipe leaves the lock and what it is told to keep, and nothing else.
func TestWipe(t *testing.T) {
	d := open(t)
	unlock, err := d.Lock()
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	for _, err := range []error{d.SaveKeys(NewKeys()), d.SaveHost(Host{Machine: "m"}), d.SaveRefused(Refused{Code: "foreign"}),
		os.Mkdir(d.file("sub"), 0o700)} {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := d.Wipe(RefusedFile); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(d.Path)
	var left []string
	for _, e := range entries {
		left = append(left, e.Name())
	}
	slices.Sort(left)
	if !slices.Equal(left, []string{LockFile, RefusedFile}) {
		t.Fatalf("left %v", left)
	}
	if err := (&Dir{Path: d.file("missing")}).Wipe(); err == nil {
		t.Fatal("wiping a missing dir succeeded")
	}
}

// The mesh lock is the process's until it lets go; another mesh's is apart.
func TestLockMesh(t *testing.T) {
	unlock, err := LockMesh("state-test-a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = LockMesh("state-test-a")
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		if err != nil {
			t.Fatalf("no lock here, yet %v", err)
		}
		unlock()
		return
	}
	if !errors.Is(err, ErrMeshLocked) {
		t.Fatalf("a second lock = %v", err)
	}
	other, err := LockMesh("state-test-b")
	if err != nil {
		t.Fatalf("another mesh's lock = %v", err)
	}
	other()
	unlock()
	again, err := LockMesh("state-test-a")
	if err != nil {
		t.Fatalf("the lock after unlock = %v", err)
	}
	again()
}

// A file of the data dir that cannot be read or removed is an error that
// names it.
func TestSessionFilesThatCannotBeUsed(t *testing.T) {
	d := open(t)
	for _, name := range []string{SessionFile, RefusedFile} {
		if err := os.MkdirAll(filepath.Join(d.file(name), "x"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := d.Session(); err == nil || !strings.Contains(err.Error(), SessionFile) {
		t.Fatalf("a session.json that is a directory = %v", err)
	}
	if err := d.ClearRefused(); err == nil || !strings.Contains(err.Error(), RefusedFile) {
		t.Fatalf("removing a refused.json that is a full directory = %v", err)
	}
}
