// Package localapi is the agent's local API (amendment A7): the CLI on the
// node (`meshctl status`, `meshctl exit`, `meshctl dns`, `meshctl logs`,
// `meshctl ui`) uses it over a Unix socket that only root and the `meshctl`
// group can open -- on Windows a named pipe only Administrators and SYSTEM
// can open. It never listens on the network. A change of the node's
// settings is relayed to the server, which decides (amendment A10); `dns
// on|off` is the node's own and stays on it.
package localapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Funcs are what the local API calls in the agent.
type Funcs struct {
	// Status reports the agent's current status.
	Status func(ctx context.Context) (proto.LocalStatus, error)
	// Prefs relays a change of the node's settings to the server.
	Prefs func(ctx context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error)
	// DNS turns the node's split DNS on or off and returns the status after.
	DNS func(ctx context.Context, on bool) (proto.LocalStatus, error)
	// Mesh returns the mesh view: fresh from the server, or the cached one.
	Mesh func(ctx context.Context) (proto.LocalMesh, error)
	// Logs is the agent's recent log.
	Logs Logs
	// UIToken mints a one-time sign-in link to the local UI.
	UIToken func(ctx context.Context) (proto.UIToken, error)
}

// Logs is the agent's recent log (logbuf.Ring).
type Logs interface {
	// Lines returns the lines held, oldest first.
	Lines() []string
	// Follow returns the lines held and a channel of the ones written
	// after; it closes when ctx ends or the follower falls behind.
	Follow(ctx context.Context) ([]string, <-chan string)
}

// Refusal is the server's answer to a relayed change, passed on as it came.
type Refusal struct {
	Status  int
	Code    string
	Message string
}

func (r *Refusal) Error() string { return r.Message }

// maxRequest caps a request body: a PrefsRequest is a few dozen bytes.
const maxRequest = 64 << 10

// Handler serves the local API.
func Handler(f Funcs) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+proto.LocalPathStatus, func(w http.ResponseWriter, r *http.Request) {
		st, err := f.Status(r.Context())
		if err != nil {
			reply(w, nil, err)
			return
		}
		reply(w, st, nil)
	})
	if f.Prefs != nil {
		mux.HandleFunc("POST "+proto.LocalPathPrefs, func(w http.ResponseWriter, r *http.Request) {
			var req proto.PrefsRequest
			if !decode(w, r, &req, "a prefs request") {
				return
			}
			// Checked here too, so a bad value never leaves the node.
			if err := req.Validate(); err != nil {
				reply(w, nil, &Refusal{Status: http.StatusBadRequest, Code: proto.CodeBadRequest, Message: err.Error()})
				return
			}
			resp, err := f.Prefs(r.Context(), req)
			reply(w, resp, err)
		})
	}
	if f.DNS != nil {
		mux.HandleFunc("POST "+proto.LocalPathDNS, func(w http.ResponseWriter, r *http.Request) {
			var req proto.DNSRequest
			if !decode(w, r, &req, "a DNS request") {
				return
			}
			st, err := f.DNS(r.Context(), req.On)
			reply(w, st, err)
		})
	}
	if f.Mesh != nil {
		mux.HandleFunc("GET "+proto.LocalPathMesh, func(w http.ResponseWriter, r *http.Request) {
			m, err := f.Mesh(r.Context())
			reply(w, m, err)
		})
	}
	if f.Logs != nil {
		mux.HandleFunc("GET "+proto.LocalPathLogs, func(w http.ResponseWriter, r *http.Request) { logs(w, r, f.Logs) })
	}
	if f.UIToken != nil {
		mux.HandleFunc("POST "+proto.LocalPathUIToken, func(w http.ResponseWriter, r *http.Request) {
			tok, err := f.UIToken(r.Context())
			reply(w, tok, err)
		})
	}
	return mux
}

// logs answers the log as text, one line per record; ?follow=true keeps
// sending new lines until the client leaves or the agent stops.
func logs(w http.ResponseWriter, r *http.Request, l Logs) {
	follow := r.URL.Query().Get("follow")
	if follow != "" {
		if _, err := validate.Enum("follow", follow, "true", "false"); err != nil {
			reply(w, nil, &Refusal{Status: http.StatusBadRequest, Code: proto.CodeBadRequest, Message: err.Error()})
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if follow != "true" {
		for _, line := range l.Lines() {
			if _, err := io.WriteString(w, line+"\n"); err != nil {
				return // the CLI went away
			}
		}
		return
	}
	held, next := l.Follow(r.Context())
	rc := http.NewResponseController(w)
	for _, line := range held {
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return
		}
	}
	if rc.Flush() != nil {
		return
	}
	for line := range next {
		if _, err := io.WriteString(w, line+"\n"); err != nil || rc.Flush() != nil {
			return
		}
	}
	if r.Context().Err() == nil {
		_, _ = io.WriteString(w, "meshctl: this reader fell behind the log and was dropped; run `meshctl logs -f` again\n") // ending anyway
	}
}

// decode reads a request body strictly, answering 400 when it is not what.
func decode(w http.ResponseWriter, r *http.Request, v any, what string) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequest))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		reply(w, nil, &Refusal{Status: http.StatusBadRequest, Code: proto.CodeBadRequest, Message: "request body is not " + what + ": " + err.Error()})
		return false
	}
	return true
}

// reply writes v, or err as a proto.ErrorBody: a Refusal with its status,
// anything else (the server unreachable) as 502.
func reply(w http.ResponseWriter, v any, err error) {
	status := http.StatusOK
	if err != nil {
		var rf *Refusal
		if !errors.As(err, &rf) {
			rf = &Refusal{Status: http.StatusBadGateway, Code: proto.CodeInternal, Message: err.Error()}
		}
		status, v = rf.Status, proto.ErrorBody{Code: rf.Code, Message: rf.Message}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v) // the CLI went away; nothing to do
}

// listenSocket opens a Unix socket at path, replacing a stale socket but
// refusing to start when another agent answers on it.
func listenSocket(path string) (net.Listener, error) {
	if len(path) > 100 {
		return nil, fmt.Errorf("local API socket path %s is too long for a Unix socket; use a shorter --socket", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", filepath.Dir(path), err)
	}
	if c, err := net.DialTimeout("unix", path, time.Second); err == nil {
		_ = c.Close()
		return nil, fmt.Errorf("another meshctl agent is already running (it answers on %s)", path)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("remove stale socket %s: %w", path, err)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", path, err)
	}
	return ln, nil
}

// Serve serves h on ln until ctx ends. Requests run in ctx, so a stream
// such as `meshctl logs -f` ends when the agent stops.
func Serve(ctx context.Context, ln net.Listener, h http.Handler) error {
	srv := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx) // ending anyway
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Client reads the local API.
type Client struct {
	socket string
	http   *http.Client
}

// NewClient returns a client for the agent on socket.
func NewClient(socket string) *Client {
	tr := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) { return dial(ctx, socket) }}
	return &Client{socket: socket, http: &http.Client{Transport: tr, Timeout: 10 * time.Second}}
}

// Status asks the agent for its status.
func (c *Client) Status(ctx context.Context) (proto.LocalStatus, error) {
	var st proto.LocalStatus
	return st, c.call(ctx, http.MethodGet, proto.LocalPathStatus, nil, &st)
}

// Prefs asks the agent to change this node's settings on the server.
func (c *Client) Prefs(ctx context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
	var resp proto.PrefsResponse
	return resp, c.call(ctx, http.MethodPost, proto.LocalPathPrefs, req, &resp)
}

// DNS asks the agent to turn the node's split DNS on or off.
func (c *Client) DNS(ctx context.Context, on bool) (proto.LocalStatus, error) {
	var st proto.LocalStatus
	return st, c.call(ctx, http.MethodPost, proto.LocalPathDNS, proto.DNSRequest{On: on}, &st)
}

// Mesh asks the agent for the mesh view.
func (c *Client) Mesh(ctx context.Context) (proto.LocalMesh, error) {
	var m proto.LocalMesh
	return m, c.call(ctx, http.MethodGet, proto.LocalPathMesh, nil, &m)
}

// UIToken asks the agent for a one-time sign-in link to the local UI.
func (c *Client) UIToken(ctx context.Context) (proto.UIToken, error) {
	var tok proto.UIToken
	return tok, c.call(ctx, http.MethodPost, proto.LocalPathUIToken, nil, &tok)
}

// Logs copies the agent's recent log to w. With follow it keeps copying
// new lines until ctx ends (nil) or the agent stops (an error).
func (c *Client) Logs(ctx context.Context, follow bool, w io.Writer) error {
	path, hc := proto.LocalPathLogs, c.http
	if follow {
		// No overall timeout: the stream is open-ended.
		path, hc = path+"?follow=true", &http.Client{Transport: c.http.Transport}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://agent"+path, nil)
	if err != nil {
		return err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return c.unreachable(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10)) // the status already failed
		return refusal(resp, b)
	}
	_, err = io.Copy(w, resp.Body)
	switch {
	case ctx.Err() != nil:
		return nil // interrupted: how `meshctl logs -f` ends
	case err != nil:
		return fmt.Errorf("read the agent's log: %w", err)
	case follow:
		return errors.New("the agent closed its log; it has stopped or is restarting")
	}
	return nil
}

func (c *Client) unreachable(err error) error {
	return fmt.Errorf("reach the agent at %s: %w (%s)", c.socket, err, reachHint)
}

// refusal is a non-200 answer as an error: the agent's ErrorBody as a
// Refusal, anything else with its status.
func refusal(resp *http.Response, body []byte) error {
	var eb proto.ErrorBody
	if json.Unmarshal(body, &eb) == nil && eb.Message != "" {
		return &Refusal{Status: resp.StatusCode, Code: eb.Code, Message: eb.Message}
	}
	return fmt.Errorf("the agent answered %s: %s", resp.Status, bytes.TrimSpace(body))
}

func (c *Client) call(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode the request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://agent"+path, body)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return c.unreachable(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read the agent's answer: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return refusal(resp, b)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode the agent's answer: %w", err)
	}
	return nil
}
