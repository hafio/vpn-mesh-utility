//go:build linux || windows

package wg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.zx2c4.com/wireguard/device"
)

// wgdev is wireguard-go as both userspace hosts run it: one device per
// agent, configured in the cross-platform protocol (UAPISet, ParseUAPI).
// The hosts differ in how they make its interface and set its addresses.
type wgdev struct {
	log *slog.Logger

	mu   sync.Mutex
	name string
	dev  *device.Device
}

// logger hands wireguard-go's log to slog: its verbose lines at debug, its
// errors (a peer that does not answer, say) as warnings.
func (w *wgdev) logger(name string) *device.Logger {
	log := w.log.With("wireguard", name)
	return &device.Logger{
		Verbosef: func(format string, args ...any) {
			if log.Enabled(context.Background(), slog.LevelDebug) {
				log.Debug(fmt.Sprintf(format, args...))
			}
		},
		Errorf: func(format string, args ...any) { log.Warn(fmt.Sprintf(format, args...)) },
	}
}

func (w *wgdev) device(name string) (*device.Device, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.dev == nil || w.name != name {
		return nil, fmt.Errorf("%s is not a WireGuard interface of this agent", name)
	}
	return w.dev, nil
}

func (w *wgdev) Device(name string) (Device, []PeerStats, error) {
	dev, err := w.device(name)
	if err != nil {
		return Device{}, nil, err
	}
	s, err := dev.IpcGet()
	if err != nil {
		return Device{}, nil, fmt.Errorf("read WireGuard device %s: %w", name, err)
	}
	return ParseUAPI(s)
}

func (w *wgdev) Configure(name string, c Change) error {
	dev, err := w.device(name)
	if err != nil {
		return err
	}
	if err := dev.IpcSet(UAPISet(c)); err != nil {
		return fmt.Errorf("configure WireGuard device %s: %w", name, err)
	}
	return nil
}
