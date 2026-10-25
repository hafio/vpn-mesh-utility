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
	// ConfigKey is the control plane's key pinned at join, and Slot the
	// node's mesh's: configs, mesh views and definitive refusals are
	// trusted only when signed with that key for this node and slot
	// (amendment A50). The join itself needs neither.
	ConfigKey ed25519.PublicKey
	Slot      proto.Slot
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
	pinned  proto.Pinned
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
		base: base, nodeID: o.NodeID, key: o.Key, ua: "meshctl/" + o.Version, now: o.Now, pinned: proto.Pinned{Key: o.ConfigKey, NodeID: o.NodeID, Slot: o.Slot},
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
	// Definitive is a refusal that ends the agent's run: a definitive code
	// whose signature verifies with the pinned key for this very request
	// (amendment A50). An unsigned or mis-signed one is not, and the agent
	// retries.
	Definitive bool
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
	status, b, _, err := c.raw(r)
	if err != nil || status == http.StatusNotModified || out == nil {
		return status, err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return status, fmt.Errorf("%s %s: decode answer: %w", r.Method, r.URL.Path, err)
	}
	return status, nil
}

// raw sends r and returns a 2xx or 304 answer as it came: status, body and
// headers. A non-2xx answer is an *Error.
func (c *Client) raw(r *http.Request) (int, []byte, http.Header, error) {
	resp, err := c.calls.Do(r)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s %s: %w", r.Method, r.URL.Path, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return resp.StatusCode, nil, nil, fmt.Errorf("%s %s: read answer: %w", r.Method, r.URL.Path, err)
	}
	if resp.StatusCode != http.StatusNotModified && resp.StatusCode/100 != 2 {
		return resp.StatusCode, nil, nil, c.refused(r, resp, b)
	}
	return resp.StatusCode, b, resp.Header, nil
}

// refused is the *Error of a non-2xx answer to r, Definitive when the
// control plane signed it for r.
func (c *Client) refused(r *http.Request, resp *http.Response, body []byte) error {
	err := apiError(resp.StatusCode, body)
	if e, ok := err.(*Error); ok && proto.Definitive(e.Code) {
		e.Definitive = c.pinned.VerifyRefusal(e.Code, r.Header.Get(proto.HeaderNonce), e.ServerTime, resp.Header.Get(proto.HeaderRefusalSignature)) == nil
	}
	return err
}

// signedBody is a 200 answer to r as the control plane signed it, checked
// as kind with the pinned key.
func (c *Client) signedBody(r *http.Request, kind string, body []byte, h http.Header) (proto.SignedBody, error) {
	b := proto.SignedBody{Body: body, Nonce: r.Header.Get(proto.HeaderNonce), Signature: h.Get(proto.HeaderConfigSignature)}
	if err := c.pinned.VerifyBody(kind, b); err != nil {
		return proto.SignedBody{}, fmt.Errorf("%s %s: %w; it was changed on the way, or comes from another control plane -- nothing of it is used",
			r.Method, r.URL.Path, err)
	}
	return b, nil
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

// Config fetches the node's desired config and checks the control plane's
// signature, returning it as signed too, for the cache. etag is the ETag of
// the applied config ("" for none); when nothing changed the server answers
// 304 and Config returns nil.
func (c *Client) Config(ctx context.Context, etag string) (*proto.NodeConfig, proto.SignedBody, error) {
	r, err := c.signed(ctx, http.MethodGet, proto.PathConfig, nil)
	if err != nil {
		return nil, proto.SignedBody{}, err
	}
	if etag != "" {
		r.Header.Set("If-None-Match", etag)
	}
	status, body, h, err := c.raw(r)
	if err != nil || status == http.StatusNotModified {
		return nil, proto.SignedBody{}, err
	}
	b, err := c.signedBody(r, proto.SignedConfig, body, h)
	if err != nil {
		return nil, proto.SignedBody{}, err
	}
	var cfg proto.NodeConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, proto.SignedBody{}, fmt.Errorf("GET %s: decode answer: %w", proto.PathConfig, err)
	}
	return &cfg, b, nil
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

// Mesh fetches the mesh view for the local UI (spec S7.6) and checks the
// control plane's signature and the view itself before anything caches or
// shows it; the view comes back as signed too, for the cache.
func (c *Client) Mesh(ctx context.Context) (proto.MeshView, proto.SignedBody, error) {
	r, err := c.signed(ctx, http.MethodGet, proto.PathMesh, nil)
	if err != nil {
		return proto.MeshView{}, proto.SignedBody{}, err
	}
	_, body, h, err := c.raw(r)
	if err != nil {
		return proto.MeshView{}, proto.SignedBody{}, err
	}
	b, err := c.signedBody(r, proto.SignedMesh, body, h)
	if err != nil {
		return proto.MeshView{}, proto.SignedBody{}, err
	}
	var v proto.MeshView
	if err := json.Unmarshal(body, &v); err != nil {
		return proto.MeshView{}, proto.SignedBody{}, fmt.Errorf("GET %s: decode answer: %w", proto.PathMesh, err)
	}
	if err := v.Validate(); err != nil {
		return proto.MeshView{}, proto.SignedBody{}, fmt.Errorf("the server's mesh view is not valid, so it is not shown: %w", err)
	}
	return v, b, nil
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
		return c.refused(r, resp, b)
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
