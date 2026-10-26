package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Host is the machine the data dir last ran on (amendment A50).
type Host struct {
	// Machine is hostid.Local of the machine's id.
	Machine string `json:"machine"`
}

// Host reads host.json; ok is false when there is none.
func (d *Dir) Host() (h Host, ok bool, err error) {
	ok, err = d.readJSON(HostFile, &h)
	return h, ok, err
}

// SaveHost writes host.json.
func (d *Dir) SaveHost(h Host) error { return d.writeJSON(HostFile, h) }

// Session is the agent's session with the control plane (amendment A50).
type Session struct {
	Session string `json:"session"`
	Ticket  string `json:"ticket"`
	// Heartbeat is the status interval the hello answered, in seconds.
	Heartbeat int `json:"heartbeat"`
	// Released is a session the agent ended with goodbye: the next start
	// says hello. The file stays as the sign that the data dir ran here.
	Released bool `json:"released,omitempty"`
}

// Session reads session.json; ok is false when there is none.
func (d *Dir) Session() (s Session, ok bool, err error) {
	if ok, err = d.readJSON(SessionFile, &s); err == nil && ok && (!proto.ValidSecret(s.Session) || !proto.ValidSecret(s.Ticket)) {
		err = fmt.Errorf("%s is damaged; the agent says hello again", d.file(SessionFile))
	}
	return s, ok, err
}

// SaveSession writes session.json.
func (d *Dir) SaveSession(s Session) error { return d.writeJSON(SessionFile, s) }

// ClearSession removes session.json.
func (d *Dir) ClearSession() error { return d.remove(SessionFile) }

// Refused is why the control plane refused this copy of the node: a code
// (proto.CodeDuplicate, CodeReplaced or CodeForeign), when and what it
// said. It never holds a key, a token or a ticket.
type Refused struct {
	Code    string `json:"code"`
	At      int64  `json:"at"`
	Message string `json:"message"`
	// Token is the fingerprint of the join token the node last joined with
	// (JoinRecord), so a container restarted with the same token is told
	// it needs a new one without asking the server again.
	Token string `json:"token_fingerprint,omitempty"`
}

// Refused reads refused.json; ok is false when there is none.
func (d *Dir) Refused() (r Refused, ok bool, err error) {
	ok, err = d.readJSON(RefusedFile, &r)
	return r, ok, err
}

// SaveRefused writes refused.json.
func (d *Dir) SaveRefused(r Refused) error { return d.writeJSON(RefusedFile, r) }

// ClearRefused removes refused.json: the node joined again, or a hello
// was granted.
func (d *Dir) ClearRefused() error { return d.remove(RefusedFile) }

// Wipe removes everything in the data dir but the lock and keep: a copy
// refused as foreign (amendment A50) keeps nothing it could run with.
func (d *Dir) Wipe(keep ...string) error {
	entries, err := os.ReadDir(d.Path)
	if err != nil {
		return fmt.Errorf("wipe %s: %w", d.Path, err)
	}
	var errs []error
	for _, e := range entries {
		if e.Name() == LockFile || slices.Contains(keep, e.Name()) {
			continue
		}
		if err := os.RemoveAll(d.file(e.Name())); err != nil {
			errs = append(errs, err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("wipe %s: %w", d.Path, err)
	}
	return nil
}

func (d *Dir) readJSON(name string, v any) (bool, error) {
	b, err := os.ReadFile(d.file(name))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("read %s: %w", d.file(name), err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return false, fmt.Errorf("%s is damaged: %w", d.file(name), err)
	}
	return true, nil
}

func (d *Dir) writeJSON(name string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	return fsutil.WriteAtomic(d.Path, name, append(b, '\n'), 0o600)
}

func (d *Dir) remove(name string) error {
	if err := os.Remove(d.file(name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", d.file(name), err)
	}
	return nil
}
