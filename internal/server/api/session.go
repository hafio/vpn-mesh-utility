package api

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// One node identity, one live connection (amendment A50). A node holds at
// most one session. Every signed request but a hello presents the session
// and the ticket its previous answer carried, and gets the next one; a
// retry after a lost answer presents the previous ticket, which is taken
// once more. Anything else means two copies of the node share the session,
// or the session is gone: 409 session, and the copy says hello. A hello
// while a live holder has the session challenges the holder and answers
// 409 wait; the holder proves it is alive by answering the challenge, and
// the waiting copy's next hello is refused, signed, as a duplicate. A
// holder that does not answer within the window loses the session to the
// waiting copy. A hello from another machine than the node's is refused,
// signed, as foreign.

// secretHash is what the store keeps of a session, a ticket or a machine
// id.
func secretHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sameHash compares two hashes in constant time.
func sameHash(a, b string) bool {
	return a != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// newSecret is a fresh session or ticket.
func newSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // never fails (crypto/rand)
	return hex.EncodeToString(b)
}

// window is how long a challenged holder has to answer, and how long a
// silent one counts as alive.
func (a *api) window() int64 {
	if a.SessionWindow > 0 {
		return int64(a.SessionWindow / time.Second)
	}
	return proto.SessionWindow
}

// challengeNonce is the nonce of the challenge open on session s: derived
// from the control plane's key, the node, the session and when the
// challenge opened, so it is the same at every delivery -- also after a
// restart -- and nobody without the key can guess it.
func (a *api) challengeNonce(node int64, s store.AgentSession) string {
	msg := "meshctl-challenge\n" + strconv.FormatInt(node, 10) + "\n" + s.SessionHash + "\n" + strconv.FormatInt(s.ChallengeAt, 10)
	sum := sha256.Sum256(ed25519.Sign(a.ConfigKey, []byte(msg)))
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum[:16])
}

// errSession sends the agent back to hello.
func errSession(n store.Node) error {
	return errorf(http.StatusConflict, proto.CodeSession,
		"node %q's session is not this agent's -- it ended, lapsed, or another copy of the node holds it; the agent says hello again", n.Name)
}

// advance checks a request's session and ticket and moves the ticket on:
// the answer carries the next one, whatever the handler answers, and the
// challenge open on the session, if any.
func (a *api) advance(ctx context.Context, w http.ResponseWriter, n store.Node, s proto.Signed) error {
	if s.Session == "" || s.Ticket == "" {
		return errSession(n)
	}
	next, now := newSecret(), a.Now().Unix()
	challenge := ""
	err := a.DB.Tx(ctx, func(tx *store.Tx) error {
		cur, err := tx.AgentSession(n.ID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			return errSession(n)
		case err != nil:
			return err
		case cur.Released || !sameHash(cur.SessionHash, secretHash(s.Session)):
			return errSession(n)
		}
		switch t := secretHash(s.Ticket); {
		case sameHash(cur.TicketHash, t):
			cur.PrevHash = cur.TicketHash // the holder's next request
		case sameHash(cur.PrevHash, t):
			// A retry after a lost answer: the ticket it never got is
			// replaced.
		default:
			return errSession(n)
		}
		cur.TicketHash, cur.LastSeen = secretHash(next), now
		if cur.ChallengeAt != 0 && !cur.Answered && a.ConfigKey != nil {
			challenge = a.challengeNonce(n.ID, cur)
		}
		return tx.PutAgentSession(cur)
	})
	if err != nil {
		return err
	}
	w.Header().Set(proto.HeaderNextTicket, next)
	if challenge != "" {
		w.Header().Set(proto.HeaderChallenge, challenge)
	}
	return nil
}

// presented checks the event stream's session and ticket, the current one
// or the previous, without moving the ticket.
func (a *api) presented(ctx context.Context, n store.Node, s proto.Signed) error {
	return a.DB.Tx(ctx, func(tx *store.Tx) error {
		cur, err := tx.AgentSession(n.ID)
		switch {
		case errors.Is(err, store.ErrNotFound):
			return errSession(n)
		case err != nil:
			return err
		case s.Session == "" || cur.Released || !sameHash(cur.SessionHash, secretHash(s.Session)):
			return errSession(n)
		}
		if t := secretHash(s.Ticket); !sameHash(cur.TicketHash, t) && !sameHash(cur.PrevHash, t) {
			return errSession(n)
		}
		return nil
	})
}

// audit records what happened to a node's connection, best effort: the
// answer does not depend on it.
func (a *api) audit(r *http.Request, n store.Node, action string, detail any) {
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		return tx.In(n.MeshID).Audit("node:"+strconv.FormatInt(n.ID, 10), 0, action, n.Name, detail)
	}); err != nil {
		a.Log.Error("audit", "action", action, "node", n.Name, "err", err)
	}
}

// hello opens a session for the node, as the package comment says.
func (a *api) hello(w http.ResponseWriter, r *http.Request) {
	n, body, err := a.authenticateAs(w, r, useHello)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var req proto.HelloRequest
	if err := decodeJSON(body, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}
	if a.ConfigKey == nil {
		a.fail(w, r, errNoConfigKey)
		return
	}
	now, machine := a.Now().Unix(), secretHash(req.MachineID)
	session, ticket := newSecret(), newSecret()
	var (
		refuse    error  // the answer when the hello is not granted
		challenge string // a challenge to deliver to the holder's stream
	)
	err = a.DB.Tx(r.Context(), func(raw *store.Tx) error {
		tx := raw.In(n.MeshID)
		actor := "node:" + strconv.FormatInt(n.ID, 10)
		m, err := tx.Machine(n.ID)
		switch {
		case errors.Is(err, store.ErrNotFound):
		case err != nil:
			return err
		case !sameHash(m.Hash, machine):
			refuse = errorf(http.StatusForbidden, proto.CodeForeign,
				"node %q runs on another machine (%s); a node stays on the machine it joined on, so this copy stops and removes its data -- "+
					"to run the node here, give it a new join token (`meshctl node token %s`)", n.Name, m.Hostname, n.Name)
			if err := tx.RecordRefusal(n.ID, store.Refusal{Code: proto.CodeForeign, Hostname: req.Hostname, At: now}); err != nil {
				return err
			}
			return tx.Audit(actor, 0, "node.foreign", n.Name, map[string]any{"hostname": req.Hostname, "machine": m.Hostname})
		}
		cur, err := tx.AgentSession(n.ID)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		lapsed := false
		if err == nil && !cur.Released && now-cur.LastSeen < a.window() {
			seen, open := now-cur.LastSeen, cur.ChallengeAt != 0 && now-cur.ChallengeAt < a.window()
			switch {
			case !open && (cur.ChallengeAt == 0 || cur.Answered):
				// None open, or the last one is over: ask the holder.
				cur.ChallengeAt, cur.Answered, cur.WaitingHost = now, false, req.Hostname
				if err := tx.PutAgentSession(cur); err != nil {
					return err
				}
				challenge = a.challengeNonce(n.ID, cur)
				refuse = errWait(n, cur, seen)
				return nil
			case open && cur.Answered:
				refuse = errorf(http.StatusForbidden, proto.CodeDuplicate,
					"node %q is running on %s and proved it is alive; this copy is a second one and stops -- stop the other copy first, "+
						"or run the node here with a new join token (`meshctl node token %s`)", n.Name, cur.Hostname, n.Name)
				cur.ChallengeAt, cur.Answered, cur.WaitingHost = 0, false, ""
				if err := tx.PutAgentSession(cur); err != nil {
					return err
				}
				if err := tx.RecordRefusal(n.ID, store.Refusal{Code: proto.CodeDuplicate, Hostname: req.Hostname, SameMachine: true, At: now}); err != nil {
					return err
				}
				return tx.Audit(actor, 0, "node.duplicate", n.Name, map[string]any{"hostname": req.Hostname, "holder": cur.Hostname})
			case open:
				challenge = a.challengeNonce(n.ID, cur) // again, in case the first was lost
				refuse = errWait(n, cur, seen)
				return nil
			}
			lapsed = true // challenged, and no answer within the window
		}
		if err := tx.PutAgentSession(store.AgentSession{NodeID: n.ID, SessionHash: secretHash(session), TicketHash: secretHash(ticket),
			LastSeen: now, Hostname: req.Hostname}); err != nil {
			return err
		}
		if err := tx.PinMachine(n.ID, store.Machine{Hash: machine, Hostname: req.Hostname, PinnedAt: now}); err != nil {
			return err
		}
		return tx.Audit(actor, 0, "node.session", n.Name, map[string]any{"hostname": req.Hostname, "lapsed": lapsed})
	})
	if challenge != "" {
		a.Hub.Challenge(n.ID, challenge)
	}
	switch {
	case err != nil:
		a.fail(w, r, err)
	case refuse != nil:
		a.fail(w, r, refuse)
	default:
		writeJSON(w, http.StatusOK, proto.HelloResponse{Session: session, Ticket: ticket, Heartbeat: proto.Heartbeat, ServerTime: now})
	}
}

// errWait answers a hello while the holder of n's session is asked
// whether it is alive.
func errWait(n store.Node, holder store.AgentSession, seen int64) error {
	return errorf(http.StatusConflict, proto.CodeWait,
		"another copy of node %q holds its session (on %s, seen %ds ago); this one waits while the control plane asks that one whether it "+
			"is still running, and touches nothing meanwhile", n.Name, holder.Hostname, seen)
}

// alive is the holder's answer to a challenge. A nonce that is not the
// open challenge's is ignored.
func (a *api) alive(w http.ResponseWriter, r *http.Request) {
	n, body, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var req proto.AliveRequest
	if err := decodeJSON(body, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}
	answered := false
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		cur, err := tx.AgentSession(n.ID)
		if err != nil || cur.ChallengeAt == 0 || cur.Answered || a.ConfigKey == nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(a.challengeNonce(n.ID, cur)), []byte(req.Nonce)) != 1 {
			return nil
		}
		cur.Answered, answered = true, true
		return tx.PutAgentSession(cur)
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	if answered {
		w.Header().Del(proto.HeaderChallenge) // the one advance put on this answer
	}
	w.WriteHeader(http.StatusNoContent)
}

// bye releases the session at once: the agent stopped cleanly, and its
// next start is granted without waiting.
func (a *api) bye(w http.ResponseWriter, r *http.Request) {
	n, _, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		cur, err := tx.AgentSession(n.ID)
		if err != nil {
			return err
		}
		cur.Released = true
		return tx.PutAgentSession(cur)
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
