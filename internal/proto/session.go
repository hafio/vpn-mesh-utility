package proto

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// One node identity, one live connection (amendment A50). An agent opens a
// session with a signed hello, which carries its host's machine id; every
// later signed request presents the session and a one-time ticket, and its
// answer carries the next ticket. Two copies of a node that share a
// session fall out of step within a heartbeat; the one sent back to hello
// waits while the control plane challenges the holder, and is refused,
// signed, when the holder answers.
const (
	PathHello = "/api/v1/agent/hello"
	PathAlive = "/api/v1/agent/alive"
	PathBye   = "/api/v1/agent/bye"
)

// Session headers. HeaderSession and HeaderTicket travel with a request and
// are covered by its signature; HeaderNextTicket and HeaderChallenge come
// back with the answer.
const (
	HeaderSession    = "X-Mesh-Session"
	HeaderTicket     = "X-Mesh-Ticket"
	HeaderNextTicket = "X-Mesh-Next-Ticket"
	// HeaderChallenge asks the holder of the session to prove it is alive
	// (POST alive with the nonce): another copy of the node said hello.
	HeaderChallenge = "X-Mesh-Challenge"
)

// Session error codes. CodeWait and CodeSession are retryable and never
// signed; CodeDuplicate and CodeForeign, like CodeReplaced, are definitive.
const (
	// CodeWait answers a hello while another copy holds the node's session
	// and is being asked whether it is alive: try again after a few
	// seconds, and touch nothing on the host meanwhile.
	CodeWait = "wait"
	// CodeSession answers a request whose session is unknown, released or
	// lapsed, or whose ticket is not the session's: say hello.
	CodeSession = "session"
	// CodeDuplicate refuses a hello while the holder of the node's session
	// proved it is alive: this copy is the second one.
	CodeDuplicate = "duplicate"
	// CodeForeign refuses a hello from a machine other than the one the
	// node joined or last ran on: a new machine needs a new join token.
	CodeForeign = "foreign"
)

// SessionWindow is how long a holder that is challenged has to answer, and
// how long a holder that stopped sending is still taken to be alive: three
// heartbeats. The dashboard calls a node late at the same age.
const SessionWindow = 90

// Heartbeat is the status interval a hello answers, in seconds; an agent
// may send more often, never less.
const Heartbeat = 30

// WaitRetry is how many seconds a hello answered CodeWait waits before it
// asks again.
const WaitRetry = 5

// EventChallenge is the SSE event that carries a challenge to the holder
// at once; its data is a ChallengeEvent.
const EventChallenge = "challenge"

// ChallengeEvent is a challenge sent over the event stream.
type ChallengeEvent struct {
	Nonce string `json:"nonce"`
}

// reSecret is a session, a ticket or a machine id as they travel: 32 bytes
// as lowercase hex.
var reSecret = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ValidSecret reports whether s has the shape of a session, a ticket or a
// machine id.
func ValidSecret(s string) bool { return reSecret.MatchString(s) }

// HelloRequest opens a session: the host the agent runs on. MachineID is
// the host's machine id as the agent derives it for the control plane (a
// hash, never the id itself); Hostname is for display only.
type HelloRequest struct {
	MachineID string `json:"machine_id"`
	Hostname  string `json:"hostname"`
}

// Validate normalises the request in place and reports every bad field.
func (r *HelloRequest) Validate() error {
	var errs []error
	if !ValidSecret(r.MachineID) {
		errs = append(errs, errors.New("machine_id must be 64 lowercase hex characters"))
	}
	if h, err := validate.Hostname("hostname", r.Hostname); err != nil {
		errs = append(errs, err)
	} else {
		r.Hostname = h
	}
	return errors.Join(errs...)
}

// HelloResponse is a granted session: its id, the first ticket and the
// status interval in seconds.
type HelloResponse struct {
	Session    string `json:"session"`
	Ticket     string `json:"ticket"`
	Heartbeat  int    `json:"heartbeat"`
	ServerTime int64  `json:"server_time"`
}

// Validate checks what the agent is about to keep.
func (r HelloResponse) Validate() error {
	if !ValidSecret(r.Session) || !ValidSecret(r.Ticket) || r.Heartbeat < 1 || r.Heartbeat > 300 {
		return fmt.Errorf("the hello answer is malformed (heartbeat %d); is the control plane an older meshctl? Upgrade it", r.Heartbeat)
	}
	return nil
}

// AliveRequest answers a challenge with its nonce.
type AliveRequest struct {
	Nonce string `json:"nonce"`
}

// Validate checks the nonce's shape.
func (r AliveRequest) Validate() error {
	if !reNonce.MatchString(r.Nonce) {
		return errors.New("nonce must be 26 base32 characters")
	}
	return nil
}

// ValidChallenge reports whether s has the shape of a challenge nonce.
func ValidChallenge(s string) bool { return reNonce.MatchString(s) }
