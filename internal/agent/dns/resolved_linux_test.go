//go:build linux

package dns

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

// What systemd-resolved's absence looks like on the bus is "unavailable",
// not a failure; a refusal says what the agent lacks. Talking to the real
// resolved would change the host, so that half is left to the hosts
// themselves (docs/operations.md, "DNS").
func TestClassifyBusErrors(t *testing.T) {
	for _, err := range []error{
		dbus.Error{Name: "org.freedesktop.DBus.Error.ServiceUnknown"},
		&dbus.Error{Name: "org.freedesktop.DBus.Error.NameHasNoOwner"},
	} {
		if got := classify(err); !errors.Is(got, ErrUnavailable) {
			t.Errorf("classify(%v) = %v, want ErrUnavailable", err, got)
		}
	}
	denied := dbus.Error{Name: "org.freedesktop.DBus.Error.AccessDenied", Body: []any{"not allowed"}}
	if got := classify(denied); errors.Is(got, ErrUnavailable) || !strings.Contains(got.Error(), "CAP_NET_ADMIN") || !errors.As(got, new(dbus.Error)) {
		t.Errorf("classify(access denied) = %v", got)
	}
	other := errors.New("timeout")
	if classify(other) != other {
		t.Error("an unknown error was rewritten")
	}
}

// An interface that does not exist fails before anything is sent to
// resolved, wherever the test runs (with a system bus or without).
func TestRegistrarNeedsTheInterface(t *testing.T) {
	r := NewRegistrar()
	ctx := context.Background()
	if _, err := r.Get(ctx, "mctest-none"); err == nil {
		t.Fatal("Get of a missing interface succeeded")
	}
	if err := r.Set(ctx, "mctest-none", Registration{}); err == nil {
		t.Fatal("Set of a missing interface succeeded")
	}
	if err := r.Revert(ctx, "mctest-none"); err == nil {
		t.Fatal("Revert of a missing interface succeeded")
	}
}
