package store

import (
	"context"
	"errors"
	"testing"
)

// A node's session, machine and last refused copy (amendment A50): read
// back as written, replaced by a later write, ErrNotFound when there is
// none, and gone with the node.
func TestAgentSessions(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	id := nodes["vps-1"].ID
	tx(t, s, func(t2 *MeshTx) error {
		if _, err := t2.AgentSession(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("no session = %v", err)
		}
		if _, err := t2.Machine(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("no machine = %v", err)
		}
		if _, err := t2.Refusal(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("no refusal = %v", err)
		}
		want := AgentSession{NodeID: id, SessionHash: "s", TicketHash: "t", PrevHash: "p", LastSeen: 10, Hostname: "vps-1", ChallengeAt: 12,
			WaitingHost: "copy"}
		if err := t2.PutAgentSession(want); err != nil {
			return err
		}
		want.TicketHash, want.Released, want.Answered = "t2", true, true
		if err := t2.PutAgentSession(want); err != nil {
			return err
		}
		if got, err := t2.AgentSession(id); err != nil || got != want {
			t.Errorf("session = %+v, %v; want %+v", got, err, want)
		}
		m := Machine{Hash: "m", Hostname: "vps-1", PinnedAt: 10}
		if err := t2.PinMachine(id, m); err != nil {
			return err
		}
		m.Hash = "m2"
		if err := t2.PinMachine(id, m); err != nil {
			return err
		}
		if got, err := t2.Machine(id); err != nil || got != m {
			t.Errorf("machine = %+v, %v", got, err)
		}
		r := Refusal{Code: "duplicate", Hostname: "copy", SameMachine: true, At: 11}
		if err := t2.RecordRefusal(id, r); err != nil {
			return err
		}
		if got, err := t2.Refusal(id); err != nil || got != r {
			t.Errorf("refusal = %+v, %v", got, err)
		}
		if err := t2.DeleteAgentSession(id); err != nil {
			return err
		}
		if _, err := t2.AgentSession(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("a deleted session = %v", err)
		}
		if err := t2.PutAgentSession(want); err != nil {
			return err
		}
		if err := t2.DeleteNode(id); err != nil {
			return err
		}
		_, e1 := t2.AgentSession(id)
		_, e2 := t2.Machine(id)
		_, e3 := t2.Refusal(id)
		if !errors.Is(e1, ErrNotFound) || !errors.Is(e2, ErrNotFound) || !errors.Is(e3, ErrNotFound) {
			t.Errorf("a removed node's session, machine and refusal: %v, %v, %v", e1, e2, e3)
		}
		return nil
	})
}

// Each accessor names what failed.
func TestAgentSessionErrors(t *testing.T) {
	s, _, _ := openTest(t)
	errBack := errors.New("roll back")
	err := s.MeshTx(context.Background(), FirstMesh, func(t2 *MeshTx) error {
		for _, table := range []string{"agent_sessions", "node_machines", "agent_refusals"} {
			if _, err := t2.exec(`DROP TABLE ` + table); err != nil {
				return err
			}
		}
		for name, err := range map[string]error{
			"read session":   func() error { _, err := t2.AgentSession(1); return err }(),
			"write session":  t2.PutAgentSession(AgentSession{NodeID: 1}),
			"end session":    t2.DeleteAgentSession(1),
			"read machine":   func() error { _, err := t2.Machine(1); return err }(),
			"pin":            t2.PinMachine(1, Machine{}),
			"read refused":   func() error { _, err := t2.Refusal(1); return err }(),
			"record refused": t2.RecordRefusal(1, Refusal{}),
		} {
			if err == nil || errors.Is(err, ErrNotFound) {
				t.Errorf("%s = %v", name, err)
			}
		}
		return errBack
	})
	if !errors.Is(err, errBack) {
		t.Fatal(err)
	}
}
