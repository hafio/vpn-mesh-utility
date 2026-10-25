//go:build windows

package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

// On Windows the data dir and everything made in it are SYSTEM's and
// Administrators' (plus an unelevated creator's), and nobody else's.
func TestDataDirIsProtectedOnWindows(t *testing.T) {
	d := open(t)
	if err := d.SaveKeys(NewKeys()); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{d.Path, filepath.Join(d.Path, WGKeyFile)} {
		sd, err := windows.GetNamedSecurityInfo(p, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		acl := sd.String()
		want := 2
		if !windows.GetCurrentProcessToken().IsElevated() {
			want = 3
		}
		// The dir's ACL is protected (D:P); a file's is inherited from it.
		if (p == d.Path) != strings.HasPrefix(acl, "D:P") || strings.Count(acl, "(A;") != want || !strings.Contains(acl, ";;;SY)") ||
			!strings.Contains(acl, ";;;BA)") || strings.Contains(acl, ";;;BU)") || strings.Contains(acl, ";;;WD)") || strings.Contains(acl, ";;;AU)") {
			t.Fatalf("%s has ACL %s", p, acl)
		}
	}
	if _, err := os.ReadFile(filepath.Join(d.Path, WGKeyFile)); err != nil {
		t.Fatalf("the creator cannot read its own key: %v", err)
	}
}
