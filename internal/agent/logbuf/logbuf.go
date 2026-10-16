// Package logbuf keeps the agent's recent log lines in memory for `meshctl
// logs` and the local UI (spec S7.6). The agent writes its text log here as
// well as to stderr (the journal, or `docker logs`), so the node's log is
// readable through the local API without access to either. Nothing is
// written to disk, and the log never carries a secret.
package logbuf

import (
	"bytes"
	"context"
	"sync"
)

// Sizes of the ring.
const (
	// DefaultLines is how many lines the agent keeps.
	DefaultLines = 2000
	// maxLine cuts a longer line, so the ring stays bounded.
	maxLine = 8 << 10
	// followBuffer is how far a follower may fall behind before it is
	// dropped; a slow reader must never hold up the agent's logging.
	followBuffer = 256
)

// Ring holds the last lines written to it and hands new ones to followers.
type Ring struct {
	mu        sync.Mutex
	lines     []string
	next      int
	full      bool
	partial   []byte
	followers map[chan string]struct{}
}

// New returns a ring that keeps the last size lines.
func New(size int) *Ring {
	return &Ring{lines: make([]string, max(size, 1)), followers: map[chan string]struct{}{}}
}

// Write adds the complete lines in p; a line without its newline yet waits
// for the rest. It never fails, so it can share a writer with stderr.
func (r *Ring) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rest := p
	for {
		i := bytes.IndexByte(rest, '\n')
		if i < 0 {
			r.partial = clip(append(r.partial, rest...))
			return len(p), nil
		}
		line := clip(append(r.partial, rest[:i]...))
		r.partial, rest = r.partial[:0], rest[i+1:]
		r.add(string(line))
	}
}

func clip(b []byte) []byte {
	if len(b) > maxLine {
		return append(b[:maxLine-3], "..."...)
	}
	return b
}

// add stores one line and passes it on. The caller holds mu.
func (r *Ring) add(line string) {
	r.lines[r.next] = line
	r.next = (r.next + 1) % len(r.lines)
	r.full = r.full || r.next == 0
	for ch := range r.followers {
		select {
		case ch <- line:
		default:
			r.drop(ch)
		}
	}
}

func (r *Ring) drop(ch chan string) {
	delete(r.followers, ch)
	close(ch)
}

// Lines returns the lines held, oldest first.
func (r *Ring) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.held()
}

func (r *Ring) held() []string {
	if !r.full {
		return append([]string(nil), r.lines[:r.next]...)
	}
	return append(append([]string(nil), r.lines[r.next:]...), r.lines[:r.next]...)
}

// Follow returns the lines held now and a channel of every line written
// after them. The channel is closed when ctx ends, or earlier when the
// follower falls followBuffer lines behind.
func (r *Ring) Follow(ctx context.Context) ([]string, <-chan string) {
	ch := make(chan string, followBuffer)
	r.mu.Lock()
	lines := r.held()
	r.followers[ch] = struct{}{}
	r.mu.Unlock()
	go func() {
		<-ctx.Done()
		r.mu.Lock()
		defer r.mu.Unlock()
		if _, ok := r.followers[ch]; ok {
			r.drop(ch)
		}
	}()
	return lines, ch
}
