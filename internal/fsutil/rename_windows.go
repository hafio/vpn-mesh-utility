//go:build windows

package fsutil

import (
	"errors"
	"syscall"
)

// errSharingViolation is ERROR_SHARING_VIOLATION.
const errSharingViolation = syscall.Errno(32)

// transientRename reports a rename that fails only while another handle has
// the target open: ERROR_ACCESS_DENIED or ERROR_SHARING_VIOLATION.
func transientRename(err error) bool {
	return errors.Is(err, syscall.ERROR_ACCESS_DENIED) || errors.Is(err, errSharingViolation)
}
