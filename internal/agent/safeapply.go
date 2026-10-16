package agent

import (
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Safe apply (spec S6.4, amendment A15) keeps a server reachable: a config
// the server pushes is applied at once but only confirmed when a status
// report sent after the apply succeeds within the window. Without that
// answer the previous config goes back on, the rollback is reported, and the
// rejected config is not applied again until the server's differs.
//
//	idle --server-pushed change--> awaiting confirmation
//	awaiting --status answered--> idle (confirmed: config.json is the new one)
//	awaiting --window passed----> idle (rolled back: the previous one again)
//
// Only server-pushed changes over a previous config arm it: a config read
// from the cache at boot, and the first config a node ever gets, have
// nothing safer to go back to.

// defaultConfirmWindow is how long a new config has to prove the server is
// still reachable.
const defaultConfirmWindow = 60 * time.Second

// pendingApply is a config applied and awaiting confirmation.
type pendingApply struct {
	cfg  proto.NodeConfig
	hash string
	// prev is the last confirmed config, what a rollback restores.
	prev     proto.NodeConfig
	armedAt  time.Time
	deadline time.Time
}

// safeApply is the state machine; it does no I/O and takes the time from
// its caller.
type safeApply struct {
	window  time.Duration
	pending *pendingApply
}

// arm starts waiting for confirmation of cfg, applied at now over prev. A
// config that replaces one still pending keeps that one's prev: only a
// confirmed config is safe to go back to.
func (s *safeApply) arm(cfg proto.NodeConfig, hash string, prev proto.NodeConfig, now time.Time) {
	if s.pending != nil {
		prev = s.pending.prev
	}
	s.pending = &pendingApply{cfg: cfg, hash: hash, prev: prev, armedAt: now, deadline: now.Add(s.window)}
}

// disarm drops a pending config without confirming it: the server's config
// went back to the confirmed one before the window ended.
func (s *safeApply) disarm() { s.pending = nil }

// confirm takes a successful status answer to a report sent at sentAt. A
// report sent before the apply proves nothing about the new config.
func (s *safeApply) confirm(sentAt time.Time) (pendingApply, bool) {
	if s.pending == nil || sentAt.Before(s.pending.armedAt) {
		return pendingApply{}, false
	}
	p := *s.pending
	s.pending = nil
	return p, true
}

// expired returns the pending config whose window ended by now.
func (s *safeApply) expired(now time.Time) (pendingApply, bool) {
	if s.pending == nil || now.Before(s.pending.deadline) {
		return pendingApply{}, false
	}
	p := *s.pending
	s.pending = nil
	return p, true
}

// confirmedHash is the hash of the config a rollback would restore, or ""
// when nothing is pending.
func (s *safeApply) confirmedHash() string {
	if s.pending == nil {
		return ""
	}
	h, _ := s.pending.prev.Hash() // it was applied, so it hashes
	return h
}
