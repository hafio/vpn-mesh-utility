package proto

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Event is one server-sent event. Data must be a single line (the payloads
// here are compact JSON); WriteEvent refuses anything else rather than
// emitting a frame the reader would split.
type Event struct {
	Name string
	Data []byte
}

// SSE event names.
const (
	EventConfig = "config"
)

// maxSSELine bounds one SSE line so a misbehaving peer cannot grow the
// reader's buffer without limit.
const maxSSELine = 64 << 10

// WriteEvent writes one event frame.
func WriteEvent(w io.Writer, e Event) error {
	if e.Name == "" || strings.ContainsAny(e.Name, "\r\n:") {
		return fmt.Errorf("sse: invalid event name %q", e.Name)
	}
	if bytes.ContainsAny(e.Data, "\r\n") {
		return fmt.Errorf("sse: event %q data must be a single line", e.Name)
	}
	_, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Name, e.Data)
	return err
}

// WriteComment writes a comment frame; proxies see traffic, readers ignore it.
func WriteComment(w io.Writer, text string) error {
	if strings.ContainsAny(text, "\r\n") {
		return errors.New("sse: comment must be a single line")
	}
	_, err := fmt.Fprintf(w, ": %s\n\n", text)
	return err
}

// WriteRetry tells the client how long to wait before reconnecting.
func WriteRetry(w io.Writer, d time.Duration) error {
	_, err := fmt.Fprintf(w, "retry: %d\n\n", d.Milliseconds())
	return err
}

// Reader parses an SSE stream into events, skipping comments and fields it
// does not use (id, retry).
type Reader struct {
	sc *bufio.Scanner
}

// NewReader wraps r.
func NewReader(r io.Reader) *Reader {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 4096), maxSSELine)
	return &Reader{sc: sc}
}

// Next returns the next complete event. It returns io.EOF when the stream ends
// cleanly between events and io.ErrUnexpectedEOF when it ends mid-event.
func (r *Reader) Next() (Event, error) {
	var (
		ev      Event
		data    [][]byte
		started bool
	)
	for r.sc.Scan() {
		line := r.sc.Text()
		if line == "" {
			if started && ev.Name == "" {
				ev.Name = "message"
			}
			if started {
				ev.Data = bytes.Join(data, []byte("\n"))
				return ev, nil
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "event":
			ev.Name = value
			started = true
		case "data":
			data = append(data, []byte(value))
			started = true
		}
	}
	if err := r.sc.Err(); err != nil {
		return Event{}, fmt.Errorf("sse: read stream: %w", err)
	}
	if started {
		return Event{}, io.ErrUnexpectedEOF
	}
	return Event{}, io.EOF
}
