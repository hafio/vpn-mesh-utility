package testfix

import (
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The fixture sessions and tickets have the shape the agent API takes, and
// differ by node and by kind.
func TestSessionFixtures(t *testing.T) {
	for _, s := range []string{Session(1), Session(2), Ticket(1), MachineID, OtherMachine} {
		if !proto.ValidSecret(s) {
			t.Errorf("%q is not a valid secret", s)
		}
	}
	if Session(1) == Session(2) || Session(1) == Ticket(1) || MachineID == OtherMachine {
		t.Fatal("fixtures collide")
	}
}
