package logbuf

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The log file keeps at most max bytes beside one rotated file, and picks up
// where it left off after a restart.
func TestFileRotates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.log")
	f, err := OpenFile(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"first line\n", "second\n", "third line\n"} {
		if _, err := io.WriteString(f, line); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	read := func(p string) string {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if got, old := read(path), read(path+".1"); got != "third line\n" || old != "first line\nsecond\n" {
		t.Fatalf("after rotation: %q, rotated %q", got, old)
	}

	again, err := OpenFile(path, 20)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(again, "fourth\n")
	_, _ = io.WriteString(again, "fifth line\n")
	_ = again.Close()
	if got, old := read(path), read(path+".1"); got != "fifth line\n" || old != "third line\nfourth\n" {
		t.Fatalf("after a restart: %q, rotated %q", got, old)
	}

	if _, err := OpenFile(filepath.Join(t.TempDir(), "missing", "agent.log"), 20); err == nil || !strings.Contains(err.Error(), "open log file") {
		t.Fatalf("OpenFile in a missing dir = %v", err)
	}

	// A rotated copy that cannot be replaced is reported, and the log goes
	// on in the file it has.
	stuck := filepath.Join(t.TempDir(), "agent.log")
	if err := os.MkdirAll(filepath.Join(stuck+".1", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	f, err = OpenFile(stuck, 20)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = io.WriteString(f, "first line\n")
	if _, err := io.WriteString(f, "second line\n"); err == nil || !strings.Contains(err.Error(), "rotate log file") {
		t.Fatalf("a write past the cap with %s.1 in the way = %v", stuck, err)
	}
	if err := f.Close(); err != nil || f.Close() != nil {
		t.Fatalf("Close = %v (twice is fine)", err)
	}
	if got := read(stuck); got != "first line\nsecond line\n" {
		t.Fatalf("the log after a failed rotation = %q", got)
	}
}
