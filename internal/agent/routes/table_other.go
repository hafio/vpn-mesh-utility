//go:build !linux && !windows

package routes

import (
	"errors"
	"net/netip"
)

type unsupported struct{}

// NewTable has no routing table to manage on this platform.
func NewTable() Table { return unsupported{} }

var errUnsupported = errors.New("route management runs on Linux and Windows only")

func (unsupported) List(string) ([]netip.Prefix, error) { return nil, errUnsupported }
func (unsupported) Add(string, netip.Prefix) error      { return errUnsupported }
func (unsupported) Del(string, netip.Prefix) error      { return errUnsupported }
