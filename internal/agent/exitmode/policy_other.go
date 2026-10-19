//go:build !linux && !windows

package exitmode

import "errors"

// errPlatform is why exit mode cannot be turned on here.
var errPlatform = errors.New("exit mode runs on Linux and Windows only")

// otherPolicy has nothing installed and cannot install anything, so exit
// mode off is a no-op and on fails with errPlatform.
type otherPolicy struct{}

// NewPolicy returns the platform's policy routing.
func NewPolicy() Policy { return otherPolicy{} }

func (otherPolicy) Rules() ([]Rule, error)   { return nil, nil }
func (otherPolicy) AddRule(Rule) error       { return errPlatform }
func (otherPolicy) DelRule(Rule) error       { return errPlatform }
func (otherPolicy) Routes() ([]Route, error) { return nil, nil }
func (otherPolicy) AddRoute(Route) error     { return errPlatform }
func (otherPolicy) DelRoute(Route) error     { return errPlatform }
func (otherPolicy) IPv6() bool               { return false }
