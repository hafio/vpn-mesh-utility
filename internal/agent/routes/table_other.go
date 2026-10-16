//go:build !linux

package routes

import (
	"errors"
	"net/netip"
)

type unsupported struct{}

// NewTable is Linux-only until the Windows backend (M5).
func NewTable() Table { return unsupported{} }

var errUnsupported = errors.New("route management on this platform arrives with the Windows agent (milestone M5)")

func (unsupported) List(string) ([]netip.Prefix, error) { return nil, errUnsupported }
func (unsupported) Add(string, netip.Prefix) error      { return errUnsupported }
func (unsupported) Del(string, netip.Prefix) error      { return errUnsupported }
