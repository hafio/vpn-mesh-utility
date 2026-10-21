package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestWriteAtomicReplacesAndLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	for _, body := range []string{"one", "two", "three"} {
		if err := WriteAtomic(dir, "x.json", []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "x.json" {
		t.Fatalf("dir holds %v, %v", entries, err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "x.json"))
	st, _ := os.Stat(filepath.Join(dir, "x.json"))
	if string(b) != "three" || (runtime.GOOS != "windows" && st.Mode().Perm() != 0o600) {
		t.Fatalf("content %q, mode %v", b, st.Mode().Perm())
	}
	if err := WriteAtomic(filepath.Join(dir, "missing"), "x", nil, 0o600); err == nil {
		t.Fatal("WriteAtomic into a missing dir succeeded")
	}
}

// A reader holding the target open does not make the write fail: on Windows
// the rename waits for it, elsewhere it never had to.
func TestWriteAtomicReplacesAFileThatIsOpen(t *testing.T) {
	dir := t.TempDir()
	if err := WriteAtomic(dir, "x.json", []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	reader, err := os.Open(filepath.Join(dir, "x.json"))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = reader.Close() // the reader is done
	}()
	if err := WriteAtomic(dir, "x.json", []byte("new"), 0o600); err != nil {
		t.Fatalf("a write while the file was open: %v", err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "x.json")); string(b) != "new" {
		t.Fatalf("content %q", b)
	}
}

// The retry: a transient failure is tried again until the way is clear or
// the budget is spent; any other failure returns at once.
func TestRenameRetrying(t *testing.T) {
	always := func(error) bool { return true }
	setup := func(t *testing.T) (tmp, target string) {
		dir := t.TempDir()
		tmp, target = filepath.Join(dir, "tmp"), filepath.Join(dir, "target")
		if err := os.WriteFile(tmp, []byte("new"), 0o600); err != nil {
			t.Fatal(err)
		}
		// A directory with a file in it cannot be renamed over: a stand-in
		// for a target that is busy.
		if err := os.MkdirAll(filepath.Join(target, "busy"), 0o700); err != nil {
			t.Fatal(err)
		}
		return tmp, target
	}

	t.Run("cleared", func(t *testing.T) {
		tmp, target := setup(t)
		go func() {
			time.Sleep(30 * time.Millisecond)
			_ = os.RemoveAll(target) // the way clears
		}()
		if err := renameRetrying(tmp, target, 5*time.Second, always); err != nil {
			t.Fatalf("rename once the target cleared: %v", err)
		}
		if b, _ := os.ReadFile(target); string(b) != "new" {
			t.Fatalf("content %q", b)
		}
	})
	t.Run("budget spent", func(t *testing.T) {
		tmp, target := setup(t)
		start := time.Now()
		if err := renameRetrying(tmp, target, 20*time.Millisecond, always); err == nil || time.Since(start) > 2*time.Second {
			t.Fatalf("rename over a busy target = %v after %s", err, time.Since(start))
		}
	})
	t.Run("not transient", func(t *testing.T) {
		tmp, target := setup(t)
		calls := 0
		err := renameRetrying(tmp, target, 5*time.Second, func(error) bool { calls++; return false })
		if err == nil || calls != 1 {
			t.Fatalf("rename = %v after %d checks, want one failure", err, calls)
		}
	})
}

func TestTransientRename(t *testing.T) {
	// Errno 5 is ERROR_ACCESS_DENIED on Windows and EIO elsewhere.
	busy := &os.LinkError{Op: "rename", Old: "a", New: "b", Err: syscall.Errno(5)}
	if got := transientRename(busy); got != (runtime.GOOS == "windows") {
		t.Fatalf("transientRename(errno 5) = %v on %s", got, runtime.GOOS)
	}
	if transientRename(os.ErrNotExist) {
		t.Fatal("a missing file is not a reader in the way")
	}
}
