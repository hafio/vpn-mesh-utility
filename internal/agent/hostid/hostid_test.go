package hostid

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A machine id file is read trimmed; a missing, empty, placeholder or
// unprintable one is refused, saying where it looked and what to mount.
func TestRead(t *testing.T) {
	dir := t.TempDir()
	write := func(name, s string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	if id, err := Read(write("ok", "4c4c4544004d3110804ac2c04f4a3332\n")); err != nil || id != "4c4c4544004d3110804ac2c04f4a3332" {
		t.Fatalf("Read = %q, %v", id, err)
	}
	for name, s := range map[string]string{"empty": "\n", "zeros": "00000000-0000-0000-0000-000000000000", "uninitialized": "uninitialized\n",
		"spaces": "abc def", "long": strings.Repeat("a", maxID+1)} {
		if _, err := Read(write(name, s)); !errors.Is(err, errNoID) {
			t.Errorf("%s: Read = %v", name, err)
		}
	}
	if _, err := Read(filepath.Join(dir, "none")); err == nil || !strings.Contains(err.Error(), "machine id") {
		t.Fatalf("a missing file = %v", err)
	}
	// The host's own: found, or the error says what to mount.
	if id, err := Read(""); err != nil && !strings.Contains(err.Error(), "/etc/machine-id:/etc/machine-id:ro") {
		t.Fatalf("the host's machine id = %q, %v", id, err)
	}
}

// What leaves the host and what the data dir keeps are two different
// hashes of the id, neither the id itself.
func TestHashes(t *testing.T) {
	const id = "4c4c4544004d3110804ac2c04f4a3332"
	p, l := Presented(id), Local(id)
	if len(p) != 64 || len(l) != 64 || p == l || strings.Contains(p+l, id) || Presented("other") == p {
		t.Fatalf("Presented %s, Local %s", p, l)
	}
}

// A host where no source holds an id says where it looked and what to
// mount; the first source that holds one wins.
func TestFirst(t *testing.T) {
	none := func() (string, error) { return "", errNoID }
	one := func() (string, error) { return "id-1", nil }
	if _, err := first([]source{{"/a", none}, {"/b", none}}); err == nil || !strings.Contains(err.Error(), "/a, /b") ||
		!strings.Contains(err.Error(), "--machine-id-file") {
		t.Fatalf("no source = %v", err)
	}
	if id, err := first([]source{{"/a", none}, {"/b", one}}); err != nil || id != "id-1" {
		t.Fatalf("first = %q, %v", id, err)
	}
}
