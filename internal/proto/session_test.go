package proto_test

import (
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A hello names a machine id of the right shape and a hostname (folded);
// its answer is a session, a ticket and a sane heartbeat; an answer to a
// challenge is a nonce (amendment A50).
func TestSessionShapes(t *testing.T) {
	r := proto.HelloRequest{MachineID: testfix.MachineID, Hostname: "VPS-1.Example.COM"}
	if err := r.Validate(); err != nil || r.Hostname != "vps-1.example.com" {
		t.Fatalf("hello %+v: %v", r, err)
	}
	for name, bad := range map[string]proto.HelloRequest{
		"no machine":          {Hostname: "vps-1"},
		"upper-case machine":  {MachineID: strings.ToUpper(testfix.MachineID), Hostname: "vps-1"},
		"a hostname with ';'": {MachineID: testfix.MachineID, Hostname: "a;b"},
	} {
		if err := bad.Validate(); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	ok := proto.HelloResponse{Session: testfix.Session(1), Ticket: testfix.Ticket(1), Heartbeat: proto.Heartbeat}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string]proto.HelloResponse{
		"no session":   {Ticket: testfix.Ticket(1), Heartbeat: 30},
		"no ticket":    {Session: testfix.Session(1), Heartbeat: 30},
		"no heartbeat": {Session: testfix.Session(1), Ticket: testfix.Ticket(1)},
		"a slow one":   {Session: testfix.Session(1), Ticket: testfix.Ticket(1), Heartbeat: 301},
	} {
		if err := bad.Validate(); err == nil || !strings.Contains(err.Error(), "Upgrade") {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := (proto.AliveRequest{Nonce: testfix.Nonce}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (proto.AliveRequest{Nonce: "short"}).Validate(); err == nil || proto.ValidChallenge("short") {
		t.Fatal("a short nonce was accepted")
	}
	for _, code := range []string{proto.CodeReplaced, proto.CodeDuplicate, proto.CodeForeign} {
		if !proto.Definitive(code) {
			t.Errorf("%s is not definitive", code)
		}
	}
	for _, code := range []string{proto.CodeWait, proto.CodeSession, proto.CodeNodeDisabled} {
		if proto.Definitive(code) {
			t.Errorf("%s is definitive", code)
		}
	}
}
