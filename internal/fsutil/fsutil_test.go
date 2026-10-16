package fsutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
