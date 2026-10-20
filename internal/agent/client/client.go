// Package client is the agent's side of the agent API (spec S6): the join
// request, then signed requests for config, events, status, prefs, the
// mesh view and the containers' names. Every signed request carries a fresh nonce, so retries never
// look like replays.
package client

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Limits on what the agent reads from the server.
const (
	requestTimeout = 30 * time.Second
	maxResponse    = 4 << 20
	// streamIdle ends an SSE stream that has been silent for longer than
	// three keepalive intervals: the connection is dead, or a proxy is
	// buffering it.
	streamIdle = 75 * time.Second
)

// Options configure a Client.
type Options struct {
	// Server is the server origin (validated).
	Server string
	NodeID int64
	Key    ed25519.PrivateKey
	// Version is the agent build, sent as the User-Agent.
	Version string
	Now     func() time.Time
	// Transport overrides the HTTP transport (tests).
	Transport http.RoundTripper
}

// Client makes the agent's API calls.
type Client struct {
	base    string
	nodeID  int64
	key     ed25519.PrivateKey
	ua      string
	now     func() time.Time
	calls   *http.Client
	streams *http.Client
}

// New validates o and returns a client.
func New(o Options) (*Client, error) {
	base, err := validate.ServerURL("server", o.Server)
	if err != nil {
		return nil, err
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	tr := o.Transport
	if tr == nil {
		tr = http.DefaultTransport
	}
	return &Client{
		base: base, nodeID: o.NodeID, key: o.Key, ua: "meshctl/" + o.Version, now: o.Now,
		calls: &http.Client{Transport: tr, Timeout: requestTimeout},
		// Streams have no overall timeout; Events watches for silence.
		streams: &http.Client{Transport: tr},
	}, nil
}

// Error is a refusal from the server. Code is one of the proto.Code*
// constants; agents branch on it, never on Message.
type Error struct {
	Status     int
	Code       string
	Message    string
	ServerTime int64
}

func (e *Error) Error() string {
	return fmt.Sprintf("server refused the request (%d %s): %s", e.Status, e.Code, e.Message)
}

// Join enrolls the node. It is the one unsigned call: the token
// authenticates it.
func Join(ctx context.Context, o Options, req proto.JoinRequest) (proto.JoinResponse, error) {
	c, err := New(o)
	if err != nil {
		return proto.JoinResponse{}, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return proto.JoinResponse{}, fmt.Errorf("encode join request: %w", err)
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+proto.PathJoin, bytes.NewReader(body))
	if err != nil {
		return proto.JoinResponse{}, fmt.Errorf("build join request: %w", err)
	}
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("User-Agent", c.ua)
	var resp proto.JoinResponse
	if _, err := c.do(hr, &resp); err != nil {
		return proto.JoinResponse{}, err
	}
	return resp, nil
}

func (c *Client) signed(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build %s %s: %w", method, path, err)
	}
	proto.Sign(c.key, c.nodeID, method, r.URL.RequestURI(), c.now(), body).Apply(r.Header)
	r.Header.Set("User-Agent", c.ua)
	if len(body) > 0 {
		r.Header.Set("Content-Type", "application/json")
	}
	return r, nil
}

// do sends r and decodes a 2xx JSON answer into out (when out is non-nil).
// It returns the status code; a non-2xx answer is an *Error.
func (c *Client) do(r *http.Request, out any) (int, error) {
	resp, err := c.calls.Do(r)
	if err != nil {
		return 0, fmt.Errorf("%s %s: %w", r.Method, r.URL.Path, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("%s %s: read answer: %w", r.Method, r.URL.Path, err)
	}
	if resp.StatusCode == http.StatusNotModified {
		return resp.StatusCode, nil
	}
	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, apiError(resp.StatusCode, b)
	}
	if out != nil {
		if err := json.Unmarshal(b, out); err != nil {
			return resp.StatusCode, fmt.Errorf("%s %s: decode answer: %w", r.Method, r.URL.Path, err)
		}
	}
	return resp.StatusCode, nil
}

func apiError(status int, body []byte) error {
	var eb proto.ErrorBody
	if err := json.Unmarshal(body, &eb); err != nil || eb.Code == "" {
		// Not our server's JSON: a proxy error page, most likely.
		return &Error{Status: status, Code: proto.CodeInternal,
			Message: fmt.Sprintf("unexpected answer %q; is the server URL right and the server running?", clip(string(body)))}
	}
	return &Error{Status: status, Code: eb.Code, Message: eb.Message, ServerTime: eb.ServerTime}
}

func clip(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

// Config fetches the node's desired config. etag is the ETag of the applied
// config ("" for none); when nothing changed the server answers 304 and
// Config returns nil.
func (c *Client) Config(ctx context.Context, etag string) (*proto.NodeConfig, error) {
	r, err := c.signed(ctx, http.MethodGet, proto.PathConfig, nil)
	if err != nil {
		return nil, err
	}
	if etag != "" {
		r.Header.Set("If-None-Match", etag)
	}
	var cfg proto.NodeConfig
	status, err := c.do(r, &cfg)
	if err != nil || status == http.StatusNotModified {
		return nil, err
	}
	return &cfg, nil
}

// Status sends a heartbeat.
func (c *Client) Status(ctx context.Context, rep proto.StatusReport) (proto.StatusResponse, error) {
	body, err := json.Marshal(rep)
	if err != nil {
		return proto.StatusResponse{}, fmt.Errorf("encode status: %w", err)
	}
	r, err := c.signed(ctx, http.MethodPost, proto.PathStatus, body)
	if err != nil {
		return proto.StatusResponse{}, err
	}
	var resp proto.StatusResponse
	if _, err := c.do(r, &resp); err != nil {
		return proto.StatusResponse{}, err
	}
	return resp, nil
}

// Prefs asks the server to change this node's own settings (amendment A10).
func (c *Client) Prefs(ctx context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return proto.PrefsResponse{}, fmt.Errorf("encode prefs: %w", err)
	}
	r, err := c.signed(ctx, http.MethodPost, proto.PathPrefs, body)
	if err != nil {
		return proto.PrefsResponse{}, err
	}
	var resp proto.PrefsResponse
	if _, err := c.do(r, &resp); err != nil {
		return proto.PrefsResponse{}, err
	}
	return resp, nil
}

// Mesh fetches the mesh view for the local UI (spec S7.6) and checks it
// before anything caches or shows it.
func (c *Client) Mesh(ctx context.Context) (proto.MeshView, error) {
	r, err := c.signed(ctx, http.MethodGet, proto.PathMesh, nil)
	if err != nil {
		return proto.MeshView{}, err
	}
	var v proto.MeshView
	if _, err := c.do(r, &v); err != nil {
		return proto.MeshView{}, err
	}
	if err := v.Validate(); err != nil {
		return proto.MeshView{}, fmt.Errorf("the server's mesh view is not valid, so it is not shown: %w", err)
	}
	return v, nil
}

// Names reports the node's container names (amendment A38).
func (c *Client) Names(ctx context.Context, req proto.NamesRequest) (proto.NamesResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return proto.NamesResponse{}, fmt.Errorf("encode names: %w", err)
	}
	r, err := c.signed(ctx, http.MethodPut, proto.PathNames, body)
	if err != nil {
		return proto.NamesResponse{}, err
	}
	var resp proto.NamesResponse
	if _, err := c.do(r, &resp); err != nil {
		return proto.NamesResponse{}, err
	}
	return resp, nil
}

// Events opens the node's SSE stream and calls fn for every config event,
// and mesh for every mesh event, until the stream ends, goes silent for
// streamIdle, or ctx is cancelled. It always returns an error saying why the
// stream ended.
func (c *Client) Events(ctx context.Context, fn func(proto.ConfigEvent), mesh func()) error {
	sctx, cancel := context.WithCancel(ctx)
	defer cancel()
	r, err := c.signed(sctx, http.MethodGet, proto.PathEvents, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Accept", "text/event-stream")
	resp, err := c.streams.Do(r)
	if err != nil {
		return fmt.Errorf("open event stream: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, maxResponse)) // best effort: the status already failed
		return apiError(resp.StatusCode, b)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		return fmt.Errorf("event stream answered with Content-Type %q; a proxy in front of the server may be rewriting it", ct)
	}
	idle := time.AfterFunc(streamIdle, cancel)
	defer idle.Stop()
	rd := proto.NewReader(&watchdog{r: resp.Body, t: idle})
	for {
		ev, err := rd.Next()
		if err != nil {
			if sctx.Err() != nil && ctx.Err() == nil {
				return fmt.Errorf("event stream was silent for %s; reconnecting", streamIdle)
			}
			return fmt.Errorf("event stream ended: %w", err)
		}
		if ev.Name == proto.EventMesh {
			mesh()
			continue
		}
		if ev.Name != proto.EventConfig {
			continue
		}
		var ce proto.ConfigEvent
		if err := json.Unmarshal(ev.Data, &ce); err != nil {
			return fmt.Errorf("event stream sent a malformed config event: %w", err)
		}
		fn(ce)
	}
}

// watchdog resets the idle timer whenever the stream delivers bytes,
// keepalive comments included.
type watchdog struct {
	r io.Reader
	t *time.Timer
}

func (w *watchdog) Read(p []byte) (int, error) {
	n, err := w.r.Read(p)
	if n > 0 {
		w.t.Reset(streamIdle)
	}
	return n, err
}
