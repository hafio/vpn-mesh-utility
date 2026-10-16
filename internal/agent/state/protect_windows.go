//go:build windows

package state

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// dirSDDL is the data dir's ACL on Windows, inherited by every file in it:
// SYSTEM (the service) and Administrators, nobody else.
const dirSDDL = "D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)"

// protect gives the data dir a protected ACL, the Windows form of mode
// 0700: the keys in it are readable by SYSTEM and Administrators only. An
// unelevated process (a test, or meshctl run without "Run as
// administrator") is added too, so it keeps access to what it made.
func protect(path string) error {
	sddl := dirSDDL
	tok := windows.GetCurrentProcessToken()
	if !tok.IsElevated() {
		u, err := tok.GetTokenUser()
		if err != nil {
			return fmt.Errorf("protect data dir %s: find this user: %w", path, err)
		}
		sddl += "(A;OICI;FA;;;" + u.User.Sid.String() + ")"
	}
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return fmt.Errorf("protect data dir %s: %w", path, err)
	}
	dacl, _, err := sd.DACL()
	if err != nil {
		return fmt.Errorf("protect data dir %s: %w", path, err)
	}
	err = windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, dacl, nil)
	if err != nil {
		return fmt.Errorf("restrict data dir %s to SYSTEM and Administrators: %w (run from an elevated prompt)", path, err)
	}
	return nil
}
