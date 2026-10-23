//go:build !linux

package containers

import (
	"errors"
	"net/netip"
)

// errPlatform is why a container network cannot be set up here.
var errPlatform = errors.New("container networks run on Linux only")

// otherNet has nothing installed and cannot install anything, so a config
// without a container network is a no-op and one with it fails.
type otherNet struct{}

// NewNet returns the platform's view of the host.
func NewNet() Net { return otherNet{} }

func (otherNet) Rules() ([]Rule, error)                           { return nil, nil }
func (otherNet) AddRule(Rule) error                               { return errPlatform }
func (otherNet) DelRule(Rule) error                               { return errPlatform }
func (otherNet) IPv6Rules() (bool, error)                         { return false, errPlatform }
func (otherNet) Blackhole(bool) (bool, error)                     { return false, nil }
func (otherNet) SetBlackhole(bool, bool) error                    { return errPlatform }
func (otherNet) LinkHolding(netip.Prefix, string) (string, error) { return "", errPlatform }
func (otherNet) Addrs(string) ([]netip.Prefix, error)             { return nil, errPlatform }
func (otherNet) AddAddr(string, netip.Prefix) error               { return errPlatform }
func (otherNet) EnsureBridge(string) (bool, error)                { return false, errPlatform }
func (otherNet) Proxies(string) ([]netip.Addr, error)             { return nil, errPlatform }
func (otherNet) AddProxy(string, netip.Addr) error                { return errPlatform }
func (otherNet) DelProxy(string, netip.Addr) error                { return errPlatform }
func (otherNet) Dropped() ([]netip.Prefix, error)                 { return nil, errPlatform }
func (otherNet) SetDropped(netip.Prefix, bool) error              { return errPlatform }
