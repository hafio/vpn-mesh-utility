// Package client is the agent's side of the agent API (spec S6): the join
// request, then signed requests for config, events, status, prefs, the
// mesh view and the containers' names. Every signed request carries a
// fresh nonce, so retries never look like replays.
//
// The node's session (amendment A50): a hello opens it, every later
// request presents it with the ticket the previous answer carried, one
// request at a time, and keeps the next ticket (Options.Keep writes it to
// the data dir). Requests wait while there is no session. An answer that
// asks whether this copy is alive -- a challenge, in a header or on the
// event stream -- is answered at once.
package client

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	// Session is the session to carry on with, from the data dir; nil, or
	// a released one, waits for Hello.
	Session *Session
	// Keep is told of every change to the session -- a hello, a next
	// ticket, a goodbye or a loss -- so the data dir holds what a restart
	// carries on with. It runs while the client's lock is held.
	Keep func(Session)
}

// Session is the node's session as the client holds it (amendment A50).
type Session struct {
	ID     string
	Ticket string
	// Heartbeat is the status interval the hello answered, in seconds.
	Heartbeat int
	// Released is a session that is no longer usable: the agent said
	// goodbye, or the control plane no longer knows it.
	Released bool
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
	keep    func(Session)

	// mu orders the requests that present the session: one at a time,
	// so a ticket is never presented after the one that followed it.
	mu sync.Mutex
	// smu guards the session, ready (closed while there is a usable one)
	// and a challenge an answer carried.
	smu       sync.Mutex
	sess      Session
	ready     chan struct{}
	lost      chan struct{}
	challenge string
	answered  string
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
	c := &Client{
		base: base, nodeID: o.NodeID, key: o.Key, ua: "meshctl/" + o.Version, now: o.Now, pinned: proto.Pinned{Key: o.ConfigKey, NodeID: o.NodeID, Slot: o.Slot},
		calls: &http.Client{Transport: tr, Timeout: requestTimeout},
		// Streams have no overall timeout; Events watches for silence.
		streams: &http.Client{Transport: tr},
		keep:    o.Keep,
		ready:   make(chan struct{}),
		lost:    make(chan struct{}),
	}
	if s := o.Session; s != nil && !s.Released {
		c.sess = *s
		close(c.ready)
	} else {
		close(c.lost)
	}
	return c, nil
}

// Lost is closed while the client has no usable session: the agent says
// hello.
func (c *Client) Lost() <-chan struct{} {
	c.smu.Lock()
	defer c.smu.Unlock()
	return c.lost
}

// ErrNoSession is a request made while the node has no session with the
// control plane: the agent is saying hello, or waiting to be let in.
var ErrNoSession = errors.New("the agent has no session with the control plane yet; it is saying hello")

// Session is the session the client holds.
func (c *Client) Session() Session {
	c.smu.Lock()
	defer c.smu.Unlock()
	return c.sess
}

// HasSession reports whether the client holds a usable session.
func (c *Client) HasSession() bool {
	c.smu.Lock()
	defer c.smu.Unlock()
	return c.sess.ID != "" && !c.sess.Released
}

// setSession takes s as the session; the caller holds smu.
func (c *Client) setSession(s Session) {
	usable := c.sess.ID != "" && !c.sess.Released
	c.sess = s
	switch next := s.ID != "" && !s.Released; {
	case next && !usable:
		close(c.ready)
		c.lost = make(chan struct{})
	case !next && usable:
		c.ready = make(chan struct{})
		close(c.lost)
	}
	if c.keep != nil {
		c.keep(s)
	}
}

// release marks the session no longer usable: requests wait for a hello.
func (c *Client) release() {
	c.smu.Lock()
	defer c.smu.Unlock()
	if c.sess.ID != "" && !c.sess.Released {
		s := c.sess
		s.Released = true
		c.setSession(s)
	}
}

// waitSession waits until there is a usable session, or ctx ends.
func (c *Client) waitSession(ctx context.Context) error {
	for {
		c.smu.Lock()
		ready, ok := c.ready, c.sess.ID != "" && !c.sess.Released
		c.smu.Unlock()
		if ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", ErrNoSession, ctx.Err())
		case <-ready:
		}
	}
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

// signed builds a signed request that presents session and ticket ("" for
// none).
func (c *Client) signed(ctx context.Context, method, path, session, ticket string, body []byte) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, method, c.base+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build %s %s: %w", method, path, err)
	}
	proto.SignSession(c.key, c.nodeID, method, r.URL.RequestURI(), c.now(), session, ticket, body).Apply(r.Header)
	r.Header.Set("User-Agent", c.ua)
	if len(body) > 0 {
		r.Header.Set("Content-Type", "application/json")
	}
	return r, nil
}

// call is a ticketed request whose 2xx JSON answer is decoded into out.
func (c *Client) call(ctx context.Context, method, path string, body []byte, wait bool, out any) error {
	r, _, b, _, err := c.ticketed(ctx, method, path, body, wait, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("%s %s: decode answer: %w", r.Method, r.URL.Path, err)
	}
	return nil
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
// headers. A non-2xx answer is an *Error. Whatever the status, the answer's
// next ticket and challenge are taken.
func (c *Client) raw(r *http.Request) (int, []byte, http.Header, error) {
	resp, err := c.calls.Do(r)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("%s %s: %w", r.Method, r.URL.Path, err)
	}
	defer resp.Body.Close()
	c.observe(r, resp.Header)
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse))
	if err != nil {
		return resp.StatusCode, nil, nil, fmt.Errorf("%s %s: read answer: %w", r.Method, r.URL.Path, err)
	}
	if resp.StatusCode != http.StatusNotModified && resp.StatusCode/100 != 2 {
		err := c.refused(r, resp, b)
		var e *Error
		if errors.As(err, &e) && e.Code == proto.CodeSession && r.Header.Get(proto.HeaderTicket) != "" {
			c.release() // say hello
		}
		return resp.StatusCode, nil, nil, err
	}
	return resp.StatusCode, b, resp.Header, nil
}

// observe takes the next ticket an answer to r carries -- the ticket never
// moves back, and only on an answer that says where it goes -- and a
// challenge to answer.
func (c *Client) observe(r *http.Request, h http.Header) {
	c.smu.Lock()
	defer c.smu.Unlock()
	if next := h.Get(proto.HeaderNextTicket); proto.ValidSecret(next) && r.Header.Get(proto.HeaderSession) == c.sess.ID && c.sess.ID != "" {
		s := c.sess
		s.Ticket = next
		c.setSession(s)
	}
	if ch := h.Get(proto.HeaderChallenge); proto.ValidChallenge(ch) && ch != c.answered {
		c.challenge = ch
	}
}

// ticketed sends a request that presents the session, waiting for one when
// wait is set, then answers a challenge its answer carried. edit adds
// headers the signature does not cover (If-None-Match).
func (c *Client) ticketed(ctx context.Context, method, path string, body []byte, wait bool,
	edit func(*http.Request)) (*http.Request, int, []byte, http.Header, error) {
	if !wait && !c.HasSession() {
		return nil, 0, nil, nil, ErrNoSession
	}
	if err := c.waitSession(ctx); err != nil {
		return nil, 0, nil, nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	r, status, b, h, err := c.send(ctx, method, path, body, edit)
	c.answerChallenge(ctx)
	return r, status, b, h, err
}

// send is one ticketed request; the caller holds mu.
func (c *Client) send(ctx context.Context, method, path string, body []byte, edit func(*http.Request)) (*http.Request, int, []byte, http.Header, error) {
	s := c.Session()
	if s.ID == "" || s.Released {
		return nil, 0, nil, nil, ErrNoSession
	}
	r, err := c.signed(ctx, method, path, s.ID, s.Ticket, body)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	if edit != nil {
		edit(r)
	}
	status, b, h, err := c.raw(r)
	return r, status, b, h, err
}

// answerChallenge answers the challenge an answer carried, if any; the
// caller holds mu. A failure is left to the next answer, which carries the
// challenge again.
func (c *Client) answerChallenge(ctx context.Context) {
	c.smu.Lock()
	nonce := c.challenge
	c.challenge = ""
	c.smu.Unlock()
	if nonce == "" {
		return
	}
	body, _ := json.Marshal(proto.AliveRequest{Nonce: nonce}) // a struct of one string always encodes
	if _, _, _, _, err := c.send(ctx, http.MethodPost, proto.PathAlive, body, nil); err == nil {
		c.smu.Lock()
		c.answered = nonce
		c.smu.Unlock()
	}
}

// Hello opens the node's session (amendment A50). A copy that another holds
// the session for gets an *Error with proto.CodeWait and should ask again
// after a few seconds; a definitive one ends its run.
func (c *Client) Hello(ctx context.Context, req proto.HelloRequest) (proto.HelloResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return proto.HelloResponse{}, fmt.Errorf("encode hello: %w", err)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	r, err := c.signed(ctx, http.MethodPost, proto.PathHello, "", "", body)
	if err != nil {
		return proto.HelloResponse{}, err
	}
	var resp proto.HelloResponse
	if _, err := c.do(r, &resp); err != nil {
		return proto.HelloResponse{}, err
	}
	if err := resp.Validate(); err != nil {
		return proto.HelloResponse{}, err
	}
	c.smu.Lock()
	c.setSession(Session{ID: resp.Session, Ticket: resp.Ticket, Heartbeat: resp.Heartbeat})
	c.challenge, c.answered = "", ""
	c.smu.Unlock()
	return resp, nil
}

// Alive answers a challenge the event stream carried.
func (c *Client) Alive(ctx context.Context, nonce string) error {
	c.smu.Lock()
	if nonce == c.answered {
		c.smu.Unlock()
		return nil
	}
	c.challenge = nonce
	c.smu.Unlock()
	if !c.HasSession() {
		return ErrNoSession
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.answerChallenge(ctx)
	return nil
}

// Bye ends the session: the agent is stopping cleanly, and its next start
// is let in at once. A goodbye the control plane did not hear keeps the
// session, which the next start carries on with.
func (c *Client) Bye(ctx context.Context) error {
	if !c.HasSession() {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, _, _, _, err := c.send(ctx, http.MethodPost, proto.PathBye, nil, nil)
	if err == nil {
		c.release()
	}
	return err
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
	r, status, body, h, err := c.ticketed(ctx, http.MethodGet, proto.PathConfig, nil, true, func(r *http.Request) {
		if etag != "" {
			r.Header.Set("If-None-Match", etag)
		}
	})
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
	var resp proto.StatusResponse
	if err := c.call(ctx, http.MethodPost, proto.PathStatus, body, true, &resp); err != nil {
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
	var resp proto.PrefsResponse
	if err := c.call(ctx, http.MethodPost, proto.PathPrefs, body, false, &resp); err != nil {
		return proto.PrefsResponse{}, err
	}
	return resp, nil
}

// Mesh fetches the mesh view for the local UI (spec S7.6) and checks the
// control plane's signature and the view itself before anything caches or
// shows it; the view comes back as signed too, for the cache.
func (c *Client) Mesh(ctx context.Context) (proto.MeshView, proto.SignedBody, error) {
	r, _, body, h, err := c.ticketed(ctx, http.MethodGet, proto.PathMesh, nil, true, nil)
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
	var resp proto.NamesResponse
	if err := c.call(ctx, http.MethodPut, proto.PathNames, body, true, &resp); err != nil {
		return proto.NamesResponse{}, err
	}
	return resp, nil
}

// Events opens the node's SSE stream and calls fn for every config event,
// and mesh for every mesh event, and answers every challenge, until the
// stream ends, goes silent for streamIdle, or ctx is cancelled. It always
// returns an error saying why the stream ended. The stream presents the
// session without moving the ticket on, so it waits for no other request.
func (c *Client) Events(ctx context.Context, fn func(proto.ConfigEvent), mesh func()) error {
	if err := c.waitSession(ctx); err != nil {
		return err
	}
	sctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s := c.Session()
	r, err := c.signed(sctx, http.MethodGet, proto.PathEvents, s.ID, s.Ticket, nil)
	if err != nil {
		return err
	}
	r.Header.Set("Accept", "text/event-stream")
	// A proxy that never answers must not hold the stream open forever.
	opening := time.AfterFunc(requestTimeout, cancel)
	resp, err := c.streams.Do(r)
	opening.Stop()
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
		if ev.Name == proto.EventChallenge {
			var ce proto.ChallengeEvent
			if json.Unmarshal(ev.Data, &ce) == nil && proto.ValidChallenge(ce.Nonce) {
				_ = c.Alive(sctx, ce.Nonce) // a failure is left to the next answer, which carries it again
			}
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
