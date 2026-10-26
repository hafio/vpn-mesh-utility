package api

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// authenticate checks a signed agent request that presents the node's
// session and ticket, and moves the ticket on (amendment A50).
func (a *api) authenticate(w http.ResponseWriter, r *http.Request) (store.Node, []byte, error) {
	return a.authenticateAs(w, r, useTicket)
}

// The ways a signed request uses the node's session (amendment A50).
const (
	// useTicket presents the session and its ticket, and moves the ticket
	// on: every request but these two.
	useTicket = iota
	// useStream presents them and leaves the ticket where it is: the event
	// stream, which lives outside the agent's one-at-a-time requests.
	useStream
	// useHello presents neither: it opens a session.
	useHello
)

// authenticateAs checks a signed agent request and returns the node and the
// body the signature covered. The cheap checks -- body size, header shape,
// clock, replay -- run before the database is touched, so a flood of junk
// costs no reads. The session is checked last, as use says.
func (a *api) authenticateAs(w http.ResponseWriter, r *http.Request, use int) (store.Node, []byte, error) {
	body, err := readBody(w, r)
	if err != nil {
		return store.Node{}, nil, err
	}
	signed, err := proto.ParseSigned(r.Header)
	if err != nil {
		return store.Node{}, nil, errorf(http.StatusUnauthorized, proto.CodeUnauthorized, "%v", err)
	}
	now := a.Now()
	if err := signed.CheckClock(now); err != nil {
		return store.Node{}, nil, errorf(http.StatusUnauthorized, proto.CodeClockSkew, "%v", err)
	}
	sig := string(signed.Signature)
	if a.replay.Seen(sig, now) {
		return store.Node{}, nil, errReplay
	}
	var (
		node    store.Node
		retired []string
	)
	err = a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		if node, err = tx.NodeByID(signed.NodeID); err != nil {
			return err
		}
		retired, err = tx.RetiredKeys(node.ID)
		return err
	})
	switch {
	case errors.Is(err, store.ErrNotFound):
		return store.Node{}, nil, errorf(http.StatusUnauthorized, proto.CodeUnauthorized,
			"node %d is not registered on this server; it was removed, or the agent's data dir belongs to another mesh -- re-join with a new token", signed.NodeID)
	case err != nil:
		return store.Node{}, nil, err
	}
	pub, err := base64.StdEncoding.DecodeString(node.IdentityPubKey)
	verr := signed.Verify(pub, r.Method, r.URL.RequestURI(), body, now)
	if !node.Enrolled || err != nil || verr != nil {
		// A key the node had before is told so, definitively (amendment
		// A50): this copy stops instead of retrying forever.
		for _, k := range retired {
			if old, err := base64.StdEncoding.DecodeString(k); err == nil && signed.Verify(old, r.Method, r.URL.RequestURI(), body, now) == nil {
				a.audit(r, node, "node.replaced", nil)
				return store.Node{}, nil, errorf(http.StatusForbidden, proto.CodeReplaced,
					"node %q no longer has this agent's key: a new join token was issued for it, or it joined again elsewhere; "+
						"this copy stops -- join it with a new token (`meshctl node token %s`) if it should run here", node.Name, node.Name)
			}
		}
	}
	switch {
	case !node.Enrolled || err != nil:
		return store.Node{}, nil, errorf(http.StatusUnauthorized, proto.CodeUnauthorized,
			"node %q has not joined yet; run `meshctl agent join` with its token", node.Name)
	case verr != nil:
		return store.Node{}, nil, errorf(http.StatusUnauthorized, proto.CodeUnauthorized,
			"%v; the agent's identity key does not match the one registered for node %q (re-join if its data dir was replaced)", verr, node.Name)
	}
	if node.Disabled {
		return store.Node{}, nil, errorf(http.StatusForbidden, proto.CodeNodeDisabled,
			"node %q is disabled; an operator can run `meshctl node enable %s`", node.Name, node.Name)
	}
	if err := a.replay.Add(sig, now); err != nil {
		return store.Node{}, nil, err
	}
	switch use {
	case useTicket:
		err = a.advance(r.Context(), w, node, signed)
	case useStream:
		err = a.presented(r.Context(), node, signed)
	}
	if err != nil {
		return store.Node{}, nil, err
	}
	return node, body, nil
}

// join enrolls a node with its one-time token (spec S6.1). The token is
// consumed in the same transaction that registers the keys and recomputes the
// mesh, so a join that fails for any reason leaves the token usable.
func (a *api) join(w http.ResponseWriter, r *http.Request) {
	ip := auth.ClientIP(r, a.TrustedProxies)
	if !a.joins.Allow(ip.String(), a.Now()) {
		a.fail(w, r, errorf(http.StatusTooManyRequests, proto.CodeRateLimited, "too many join attempts from %s; wait a minute", ip))
		return
	}
	body, err := readBody(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var req proto.JoinRequest
	if err := decodeJSON(body, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}

	// The token's node says which mesh the join reconciles (amendment A49).
	var mesh, id int64
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		mesh, err = tx.TokenMesh(req.Token)
		return err
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	_, err = a.Reconciler.Apply(r.Context(), Change{Mesh: mesh, Action: "node.join"}, func(tx *store.MeshTx, c *Change) error {
		var err error
		if id, err = tx.ConsumeToken(req.Token); err != nil {
			return err
		}
		n, err := tx.NodeByID(id)
		if err != nil {
			return err
		}
		switch {
		case req.Name != "" && req.Name != n.Name:
			return errorf(http.StatusConflict, proto.CodeConflict, "this token belongs to node %q, not %q; drop --name or use that node's token", n.Name, req.Name)
		case n.Disabled:
			return errorf(http.StatusForbidden, proto.CodeNodeDisabled, "node %q is disabled; run `meshctl node enable %s` first", n.Name, n.Name)
		case n.Role == string(proto.RolePhone):
			return errorf(http.StatusBadRequest, proto.CodeBadRequest, "node %q is a phone; phones join by QR code, not with an agent", n.Name)
		}
		rejoin := n.Enrolled
		// Every join brings fresh keys (amendment A50): one this node, or
		// any other, ever had is refused, and the key it replaces joins
		// the node's key history.
		switch used, err := tx.KeyRetired(req.IdentityPublicKey); {
		case err != nil:
			return err
		case used:
			return errorf(http.StatusConflict, proto.CodeConflict,
				"this identity key was cut off before; the agent makes fresh keys at every join -- is it an older meshctl? Upgrade it")
		}
		if n.IdentityPubKey != req.IdentityPublicKey {
			if err := tx.RetireKey(n.ID, n.IdentityPubKey, a.Now().Unix()); err != nil {
				return err
			}
		}
		n.WGPubKey, n.IdentityPubKey = req.WGPublicKey, req.IdentityPublicKey
		n.OS, n.Arch, n.Hostname, n.WGBackend, n.AgentVersion = req.OS, req.Arch, req.Hostname, req.WGBackend, req.AgentVersion
		n.Enrolled = true
		// A family the node does not use is left as it is (M8).
		if n.Role == string(proto.RoleServer) && n.Family != proto.FamilyIPv6 {
			n.Endpoint = pickEndpoint(req, n.Endpoint, ip, n.ListenPort)
		}
		if n.Role == string(proto.RoleServer) && n.Family != proto.FamilyIPv4 {
			n.Endpoint6 = pickEndpoint6(req, n.Endpoint6, ip, n.ListenPort)
		}
		if err := tx.UpdateNode(n); err != nil {
			if errors.Is(err, store.ErrConflict) {
				return errorf(http.StatusConflict, proto.CodeConflict,
					"another node already uses this WireGuard or identity key; was this data dir copied from another node? Remove its keys and join again")
			}
			return err
		}
		// A join pins the node to its machine and ends the session its
		// earlier keys held (amendment A50).
		if err := tx.PinMachine(n.ID, store.Machine{Hash: secretHash(req.MachineID), Hostname: req.Hostname, PinnedAt: a.Now().Unix()}); err != nil {
			return err
		}
		if err := tx.DeleteAgentSession(n.ID); err != nil {
			return err
		}
		c.Actor = "node:" + strconv.FormatInt(n.ID, 10)
		c.Target = n.Name
		detail := map[string]any{"hostname": n.Hostname, "os": n.OS, "arch": n.Arch, "wg_backend": n.WGBackend,
			"endpoint": n.Endpoint, "agent_version": n.AgentVersion, "rejoin": rejoin, "from": ip.String()}
		if n.Endpoint6 != "" {
			detail["endpoint6"] = n.Endpoint6
		}
		c.Detail = detail
		return nil
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var (
		n store.Node
		m store.Mesh
	)
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		if n, err = tx.NodeByID(id); err != nil {
			return err
		}
		m, err = tx.MeshByID(n.MeshID)
		return err
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	if a.ConfigKey == nil {
		a.fail(w, r, errNoConfigKey)
		return
	}
	writeJSON(w, http.StatusOK, proto.JoinResponse{
		NodeID: n.ID, Name: n.Name, Role: proto.Role(n.Role), MeshIP: n.MeshIP,
		ConfigVersion: n.ConfigVersion, ServerTime: a.Now().Unix(), MeshIP6: n.MeshIP6,
		Mesh: m.Name, MeshUUID: m.UUID, Slot: m.Slot,
		ConfigKey: base64.StdEncoding.EncodeToString(a.ConfigKey.Public().(ed25519.PublicKey)),
	})
}

// pickEndpoint chooses a server's public endpoint (amendment A17): the
// agent's --endpoint, else the public IPv4 it found on its interfaces, else
// what the operator set with `node add --endpoint`, else the address the
// request came from when that is public. Detected and observed addresses use
// the node's listen port.
func pickEndpoint(req proto.JoinRequest, current string, observed netip.Addr, port int) string {
	public := func(a netip.Addr) bool { return a.Is4() && a.IsGlobalUnicast() && !a.IsPrivate() }
	detected, _ := netip.ParseAddr(req.PublicIP)
	switch {
	case req.Endpoint != "":
		return req.Endpoint
	case public(detected):
		return netip.AddrPortFrom(detected, uint16(port)).String()
	case current != "":
		return current
	case public(observed):
		return netip.AddrPortFrom(observed, uint16(port)).String()
	}
	return ""
}

// pickEndpoint6 chooses a server's public IPv6 endpoint (M8) as
// pickEndpoint does the IPv4 one: the agent's --endpoint6, else the public
// IPv6 it found, else the operator's, else the address the request came
// from when that is a public IPv6.
func pickEndpoint6(req proto.JoinRequest, current string, observed netip.Addr, port int) string {
	detected, _ := netip.ParseAddr(req.PublicIP6)
	switch {
	case req.Endpoint6 != "":
		return req.Endpoint6
	case public6(detected):
		return netip.AddrPortFrom(detected, uint16(port)).String()
	case current != "":
		return current
	case public6(observed):
		return netip.AddrPortFrom(observed, uint16(port)).String()
	}
	return ""
}

// public6 is a global IPv6 address: not unique local, link local or an
// IPv4 one in disguise.
func public6(a netip.Addr) bool {
	return a.Is6() && !a.Is4In6() && a.IsGlobalUnicast() && !a.IsPrivate()
}

// config returns the node's full desired state. If-None-Match with the
// applied config's ETag answers 304, so the 60 s poll costs one hash.
func (a *api) config(w http.ResponseWriter, r *http.Request) {
	node, _, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var cfg *proto.NodeConfig
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		_, cfg, err = configFor(tx.In(node.MeshID), node.ID)
		return err
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	if cfg == nil {
		a.fail(w, r, errorf(http.StatusNotFound, proto.CodeNotFound, "node %q has no computed config", node.Name))
		return
	}
	hash, err := cfg.Hash()
	if err != nil {
		a.fail(w, r, err)
		return
	}
	etag := proto.ETag(hash)
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	a.writeSigned(w, r, proto.SignedConfig, node.ID, cfg.Node.Slot, cfg)
}

// events is the node's SSE stream (spec S6.3). It opens with a retry hint and
// the node's current version, so a reconnecting agent catches up without
// polling, then sends one `config` event per change, a `mesh` event when the
// mesh view changed (amendment A38) and a comment every KeepAlive so
// proxies keep the connection open.
func (a *api) events(w http.ResponseWriter, r *http.Request) {
	node, _, err := a.authenticateAs(w, r, useStream)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	// Subscribe before reading the current state: a change committed in
	// between is then queued rather than lost.
	events, mesh, challenge, done, cancel := a.Hub.Subscribe(node.ID, node.MeshID)
	defer cancel()
	open := ""
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		if node, err = tx.NodeByID(node.ID); err != nil {
			return err
		}
		// A challenge already open is sent at once (amendment A50).
		s, err := tx.AgentSession(node.ID)
		if err == nil && s.ChallengeAt != 0 && !s.Answered {
			open = a.challengeNonce(node.ID, s)
		}
		return nil
	}); err != nil {
		a.fail(w, r, err)
		return
	}

	rc := http.NewResponseController(w)
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	send := func(write func() error) bool {
		// Wall-clock deadline: the injected clock may be a test's fake.
		_ = rc.SetWriteDeadline(time.Now().Add(writeTimeout)) // unsupported on some writers; then the server's own timeouts apply
		return write() == nil && rc.Flush() == nil
	}
	sendConfig := func(ev proto.ConfigEvent) bool {
		data, err := json.Marshal(ev)
		if err != nil {
			return false
		}
		return send(func() error { return proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: data}) })
	}
	if !send(func() error { return proto.WriteRetry(w, sseRetry) }) {
		return
	}
	sendChallenge := func(nonce string) bool {
		data, err := json.Marshal(proto.ChallengeEvent{Nonce: nonce})
		if err != nil {
			return false
		}
		return send(func() error { return proto.WriteEvent(w, proto.Event{Name: proto.EventChallenge, Data: data}) })
	}
	if node.ConfigHash != "" && !sendConfig(proto.ConfigEvent{Version: node.ConfigVersion, Hash: node.ConfigHash}) {
		return
	}
	if open != "" && !sendChallenge(open) {
		return
	}
	ka := time.NewTicker(a.KeepAlive)
	defer ka.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-done:
			return
		case ev := <-events:
			if !sendConfig(ev) {
				return
			}
		case <-mesh:
			if !send(func() error { return proto.WriteEvent(w, proto.Event{Name: proto.EventMesh, Data: []byte("{}")}) }) {
				return
			}
		case nonce := <-challenge:
			if !sendChallenge(nonce) {
				return
			}
		case <-ka.C:
			if !send(func() error { return proto.WriteComment(w, "ka") }) {
				return
			}
		}
	}
}

// statusDetail is the display-only part of a heartbeat, kept as JSON.
type statusDetail struct {
	Errors   []string        `json:"errors,omitempty"`
	Degraded []string        `json:"degraded,omitempty"`
	Rollback *proto.Rollback `json:"rollback,omitempty"`
	DNS      string          `json:"dns,omitempty"`
}

// status takes a heartbeat (spec S6.3). The response carries the node's
// current version, so an agent that missed an event learns it here.
func (a *api) status(w http.ResponseWriter, r *http.Request) {
	node, body, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var rep proto.StatusReport
	if err := decodeJSON(body, &rep); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := rep.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}
	// The families the host can use change the plan (M8), so a new set goes
	// through the reconciler; the heartbeat itself does not.
	if fams := hostFamilies(rep.Families); fams != node.HostFamilies {
		actor := "node:" + strconv.FormatInt(node.ID, 10)
		ch := Change{Mesh: node.MeshID, Actor: actor, Action: "node.families", Target: node.Name, Detail: map[string]any{"families": fams}}
		set := func(tx *store.MeshTx, _ *Change) error { return tx.SetHostFamilies(node.ID, fams) }
		if _, err := a.Reconciler.Apply(r.Context(), ch, set); err != nil {
			a.fail(w, r, err)
			return
		}
	}
	detail, err := json.Marshal(statusDetail{Errors: rep.Errors, Degraded: rep.Degraded, Rollback: rep.Rollback, DNS: rep.DNS})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var resp proto.StatusResponse
	err = a.DB.Tx(r.Context(), func(raw *store.Tx) error {
		tx := raw.In(node.MeshID)
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		ids := make(map[string]int64, len(nodes))
		for _, n := range nodes {
			ids[n.Name] = n.ID
		}
		u := store.StatusUpdate{
			AppliedVersion: rep.AppliedVersion, AppliedHash: rep.AppliedHash, AgentVersion: rep.AgentVersion,
			WGBackend: rep.WGBackend, ObservedIP: auth.ClientIP(r, a.TrustedProxies).String(), StatusJSON: string(detail),
		}
		for _, p := range rep.Peers {
			// A peer the server no longer knows (removed a moment ago) is
			// skipped, not an error: the next config drops it.
			if id, ok := ids[p.Name]; ok && id != node.ID {
				u.Peers = append(u.Peers, store.PeerStat{PeerID: id, LastHandshake: p.LastHandshake,
					RxBytes: p.RxBytes, TxBytes: p.TxBytes, Endpoint: p.Endpoint})
			}
		}
		if err := tx.RecordStatus(node.ID, u); err != nil {
			return err
		}
		if rb := rep.Rollback; rb != nil && !sameRollback(node.StatusJSON, rb) {
			if err := tx.Audit("node:"+strconv.FormatInt(node.ID, 10), 0, "node.rollback", node.Name,
				map[string]any{"from": rb.From, "to": rb.To, "reason": rb.Reason}); err != nil {
				return err
			}
		}
		cur, err := tx.NodeByID(node.ID)
		if err != nil {
			return err
		}
		resp = proto.StatusResponse{ConfigVersion: cur.ConfigVersion, ConfigHash: cur.ConfigHash, ServerTime: a.Now().Unix()}
		return nil
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// hostFamilies is a reported family list as it is stored: "ipv4", "ipv6",
// "ipv4,ipv6", or "" when the agent did not say.
func hostFamilies(fs []string) string {
	sorted := slices.Clone(fs)
	slices.Sort(sorted) // ipv4 before ipv6
	return strings.Join(sorted, ",")
}

// sameRollback reports whether the node's previous heartbeat already carried
// this rollback: the agent repeats it until a newer config is confirmed, and
// it is audited once, not every 30 s.
func sameRollback(prevStatus string, rb *proto.Rollback) bool {
	var d statusDetail
	return json.Unmarshal([]byte(prevStatus), &d) == nil && d.Rollback != nil && *d.Rollback == *rb
}

// prefs applies a node's own settings (amendment A10) through Nodes.Set, so
// the change is versioned, audited as the node and pushed exactly like an
// operator's `meshctl node set`. Only a spoke chooses an exit or a keepalive.
func (a *api) prefs(w http.ResponseWriter, r *http.Request) {
	node, body, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var req proto.PrefsRequest
	if err := decodeJSON(body, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}
	if node.Role != string(proto.RoleSpoke) && (req.Exit != nil || req.Keepalive != nil) {
		a.fail(w, r, errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"node %q is a %s; only spokes choose an exit or a keepalive", node.Name, node.Role))
		return
	}
	ch, err := a.nodes.In(node.MeshID).Set(r.Context(), Actor{Name: "node:" + strconv.FormatInt(node.ID, 10)}, node.Name,
		NodeSetRequest{Exit: req.Exit, Keepalive: req.Keepalive, MTU: req.MTU})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var cur store.Node
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		cur, err = tx.NodeByID(node.ID)
		return err
	}); err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, proto.PrefsResponse{Version: ch.Version, ConfigVersion: cur.ConfigVersion, ConfigHash: cur.ConfigHash})
}

// mesh is the mesh view for the node's local UI (spec S7.6): every node as
// the dashboard shows it, without keys, endpoints, hostnames or errors, and
// the containers' published names, which every node's resolver answers.
func (a *api) mesh(w http.ResponseWriter, r *http.Request) {
	node, _, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	now := a.Now()
	view := proto.MeshView{ServerTime: now.Unix(), Nodes: []proto.MeshNode{}}
	var m store.Mesh
	err = a.DB.Tx(r.Context(), func(raw *store.Tx) error {
		tx := raw.In(node.MeshID)
		var err error
		if m, err = tx.MeshByID(node.MeshID); err != nil {
			return err
		}
		if view.Version, err = tx.LatestVersion(); err != nil {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		for _, n := range nodes {
			info := nodeInfo(n, nodes)
			if err := withHandshake(tx, &info); err != nil {
				return err
			}
			health, state := Health(info, now)
			view.Nodes = append(view.Nodes, proto.MeshNode{Name: info.Name, Role: proto.Role(info.Role), MeshIP: info.MeshIP,
				Hub: info.Hub, Exit: info.Exit, CanExit: info.CanExit, Health: health, State: state, LastSeen: info.Seen(),
				ConfigVersion: info.ConfigVersion, AppliedVersion: info.AppliedVersion})
		}
		names, err := a.publishedNames(tx)
		view.Names = names.Published
		return err
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	a.writeSigned(w, r, proto.SignedMesh, node.ID, proto.Slot(m.Slot), view)
}

// health answers the container healthcheck: the process is serving and the
// database answers.
func (a *api) health(w http.ResponseWriter, r *http.Request) {
	if err := a.DB.Tx(r.Context(), func(tx *store.Tx) error {
		_, err := tx.SchemaVersion()
		return err
	}); err != nil {
		a.fail(w, r, fmt.Errorf("health: %w", err))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}
