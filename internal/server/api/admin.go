package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Node administration (amendment A9). Nodes holds the operations; the admin
// API on <data>/admin.sock (mode 0600, so only the server's own user reaches
// it: `docker exec`) and the web UI both call them, so a node changes the
// same way whoever changes it. Every write goes through the reconciler and is
// audited under its Actor.

const actorCLI = "cli"

// cli is the actor of every admin socket write.
func cli() Actor { return Actor{Name: actorCLI} }

// Actor is who makes a change: "cli" for the admin socket, "user:<name>" with
// the user id for the web UI.
type Actor struct {
	Name   string
	UserID int64
}

// Nodes is the node administration service.
type Nodes struct {
	db        DB
	rec       *Reconciler
	hub       *Hub
	serverURL string
	// version is the server's build: a spoke's container runs the same
	// release.
	version string
}

// NewNodes wires the service from the API's dependencies.
func NewNodes(d Deps) *Nodes {
	return &Nodes{db: d.DB, rec: d.Reconciler, hub: d.Hub, serverURL: d.ServerURL, version: d.Version}
}

// ErrorMessage is how a failure from Nodes (or the reconciler) is shown to an
// operator: its HTTP status and a message that never carries a secret.
func ErrorMessage(err error) (int, string) {
	ae := toAPIError(err)
	return ae.status, ae.msg
}

func (a *api) adminRoutes() []route {
	const one = AdminNodes + "/{name}"
	return []route{
		{Route{http.MethodGet, AdminNodes, "socket"}, a.adminList},
		{Route{http.MethodPost, AdminNodes, "socket"}, a.adminAdd},
		{Route{http.MethodGet, one, "socket"}, a.adminShow},
		{Route{http.MethodPatch, one, "socket"}, a.adminSet},
		{Route{http.MethodDelete, one, "socket"}, a.adminRemove},
		{Route{http.MethodPost, one + "/disable", "socket"}, a.adminDisable},
		{Route{http.MethodPost, one + "/enable", "socket"}, a.adminEnable},
		{Route{http.MethodPost, one + "/token", "socket"}, a.adminToken},
		{Route{http.MethodGet, one + "/compose", "socket"}, a.adminCompose},
		{Route{http.MethodPost, one + "/phone", "socket"}, a.adminPhone},
	}
}

// reply writes v, or the failure.
func (a *api) reply(w http.ResponseWriter, r *http.Request, v any, err error) {
	if err != nil {
		a.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (a *api) adminList(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.List(r.Context())
	a.reply(w, r, out, err)
}

func (a *api) adminShow(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.Show(r.Context(), r.PathValue("name"))
	a.reply(w, r, out, err)
}

func (a *api) adminAdd(w http.ResponseWriter, r *http.Request) {
	var req NodeAddRequest
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.Add(r.Context(), cli(), req)
	a.reply(w, r, out, err)
}

func (a *api) adminSet(w http.ResponseWriter, r *http.Request) {
	var req NodeSetRequest
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.Set(r.Context(), cli(), r.PathValue("name"), req)
	a.reply(w, r, out, err)
}

func (a *api) adminDisable(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.SetDisabled(r.Context(), cli(), r.PathValue("name"), true)
	a.reply(w, r, out, err)
}

func (a *api) adminEnable(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.SetDisabled(r.Context(), cli(), r.PathValue("name"), false)
	a.reply(w, r, out, err)
}

func (a *api) adminRemove(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.Remove(r.Context(), cli(), r.PathValue("name"))
	a.reply(w, r, out, err)
}

func (a *api) adminToken(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.Token(r.Context(), cli(), r.PathValue("name"))
	a.reply(w, r, out, err)
}

func (a *api) adminCompose(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	out, err := a.nodes.Compose(r.Context(), r.PathValue("name"), q.Get("backend"), q.Get("data"))
	a.reply(w, r, out, err)
}

func (a *api) readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	body, err := readBody(w, r)
	if err != nil {
		return err
	}
	return decodeJSON(body, v)
}

// Compose renders the compose file that runs a spoke's agent as a container
// (spec S12, amendments A13 and A19): backend userspace (the default) or
// kernel, the agent's data in dataDir on the host (default
// deploy.DefaultNASData). Servers install with install.sh instead.
func (s *Nodes) Compose(ctx context.Context, name, backend, dataDir string) (ComposeFile, error) {
	name, err := checkName(name)
	if err != nil {
		return ComposeFile{}, err
	}
	var n store.Node
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		var err error
		n, err = nodeByName(tx, name, "node")
		return err
	})
	if err != nil {
		return ComposeFile{}, err
	}
	if n.Role != string(proto.RoleSpoke) {
		return ComposeFile{}, errorf(http.StatusConflict, proto.CodeConflict,
			"%s is a %s; only spokes run in a container (a server installs with the command `meshctl node add` printed)", name, n.Role)
	}
	if backend == "" {
		backend = deploy.NASUserspace
	}
	if dataDir == "" {
		dataDir = deploy.DefaultNASData
	}
	b, err := deploy.RenderNASCompose(deploy.NASParams{Node: name, Server: s.serverURL, Version: s.version, Backend: backend, DataDir: dataDir})
	if err != nil {
		return ComposeFile{}, invalid(err)
	}
	return ComposeFile{Backend: backend, DataDir: dataDir, YAML: string(b)}, nil
}

// List returns every node and the plan's problems.
func (s *Nodes) List(ctx context.Context) (NodeList, error) {
	var out NodeList
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		out.Nodes = make([]NodeInfo, 0, len(nodes))
		for _, n := range nodes {
			info := nodeInfo(n, nodes)
			if err := withHandshake(tx, &info); err != nil {
				return err
			}
			out.Nodes = append(out.Nodes, info)
		}
		out.Warnings = problemStrings(planner.Check(in))
		return nil
	})
	return out, err
}

// Show returns one node, its computed config and its last peer report.
func (s *Nodes) Show(ctx context.Context, name string) (NodeDetail, error) {
	name, err := checkName(name)
	if err != nil {
		return NodeDetail{}, err
	}
	var out NodeDetail
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		out.Node = nodeInfo(n, nodes)
		if err := withHandshake(tx, &out.Node); err != nil {
			return err
		}
		for _, p := range planner.Check(in) {
			if p.Node == "" || p.Node == n.Name {
				out.Warnings = append(out.Warnings, p.String())
			}
		}
		if _, out.Config, err = configFor(tx, n.ID); err != nil {
			out.Warnings = append(out.Warnings, fmt.Sprintf("error: no config could be computed: %v", err))
		}
		stats, err := tx.PeerStats(n.ID)
		if err != nil {
			return err
		}
		names := map[int64]string{}
		for _, o := range nodes {
			names[o.ID] = o.Name
		}
		for _, st := range stats {
			out.Peers = append(out.Peers, proto.PeerStatus{Name: names[st.PeerID], LastHandshake: st.LastHandshake,
				RxBytes: st.RxBytes, TxBytes: st.TxBytes, Endpoint: st.Endpoint})
		}
		return nil
	})
	return out, err
}

// Add creates a node and its first join token; a phone gets its profile
// instead (Phone).
func (s *Nodes) Add(ctx context.Context, by Actor, req NodeAddRequest) (NodeChange, error) {
	n, fixedIP, err := newNode(req)
	if err != nil {
		return NodeChange{}, err
	}
	var enroll *Enrollment
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.add", Target: n.Name}, func(tx *store.Tx, c *Change) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		used := map[netip.Addr]bool{}
		for _, o := range nodes {
			if ip, err := netip.ParseAddr(o.MeshIP); err == nil {
				used[ip] = true
			}
		}
		if fixedIP.IsValid() {
			if err := planner.CheckRequested(in.MeshCIDR, fixedIP, used); err != nil {
				return invalid(err)
			}
			n.MeshIP = fixedIP.String()
		} else {
			ip, err := planner.Allocate(in.MeshCIDR, planner.RangeFor(proto.Role(n.Role), req.AlwaysOn), used)
			if err != nil {
				return errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
			}
			n.MeshIP = ip.String()
		}
		if n.MTU == 0 {
			if n.MTU, err = defaultMTU(tx); err != nil {
				return err
			}
		}
		if req.Hub != "" {
			hub, err := nodeByName(tx, req.Hub, "hub")
			if err != nil {
				return err
			}
			n.PrimaryHubID = hub.ID
		}
		if err := tx.InsertNode(&n); err != nil {
			if errors.Is(err, store.ErrConflict) {
				return errorf(http.StatusConflict, proto.CodeConflict, "a node named %q already exists; pick another name or remove that node first", n.Name)
			}
			return err
		}
		c.Detail = map[string]any{"role": n.Role, "mesh_ip": n.MeshIP, "endpoint": n.Endpoint, "hub": req.Hub, "can_exit": n.CanExit}
		if n.Role == string(proto.RolePhone) {
			return nil
		}
		token, exp, err := tx.IssueToken(n.ID, store.TokenTTL)
		if err != nil {
			return err
		}
		enroll = s.enrollment(token, exp.Unix())
		return nil
	})
	if err != nil {
		return NodeChange{}, err
	}
	return s.answer(ctx, n.Name, res, enroll)
}

// Set changes the given fields of a node.
func (s *Nodes) Set(ctx context.Context, by Actor, name string, req NodeSetRequest) (NodeChange, error) {
	name, err := checkName(name)
	if err != nil {
		return NodeChange{}, err
	}
	if err := validateSet(req); err != nil {
		return NodeChange{}, err
	}
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.set", Target: name}, func(tx *store.Tx, c *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		detail := map[string]any{}
		if req.Endpoint != nil {
			n.Endpoint = *req.Endpoint
			if n.Endpoint != "" {
				ap, _ := validate.Endpoint("endpoint", n.Endpoint) // validated in validateSet
				n.Endpoint = ap.String()
			}
			detail["endpoint"] = n.Endpoint
		}
		if req.ListenPort != nil {
			n.ListenPort = *req.ListenPort
			detail["listen_port"] = n.ListenPort
		}
		if req.Hub != nil {
			if n.PrimaryHubID, err = refID(tx, *req.Hub, "hub"); err != nil {
				return err
			}
			detail["hub"] = *req.Hub
		}
		if req.CanExit != nil {
			n.CanExit = *req.CanExit
			detail["can_exit"] = n.CanExit
		}
		if req.Exit != nil {
			if n.ExitVia, err = refID(tx, *req.Exit, "exit"); err != nil {
				return err
			}
			detail["exit"] = *req.Exit
		}
		if req.Keepalive != nil {
			n.Keepalive = *req.Keepalive
			detail["keepalive"] = n.Keepalive
		}
		if req.MTU != nil {
			n.MTU = *req.MTU
			detail["mtu"] = n.MTU
		}
		c.Detail = detail
		return tx.UpdateNode(n)
	})
	if err != nil {
		return NodeChange{}, err
	}
	return s.answer(ctx, name, res, nil)
}

// SetDisabled disables a node (every peer drops it; its stream is closed) or
// enables it again.
func (s *Nodes) SetDisabled(ctx context.Context, by Actor, name string, disabled bool) (NodeChange, error) {
	name, err := checkName(name)
	if err != nil {
		return NodeChange{}, err
	}
	action := "node.enable"
	if disabled {
		action = "node.disable"
	}
	var id int64
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: action, Target: name}, func(tx *store.Tx, _ *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		id, n.Disabled = n.ID, disabled
		return tx.UpdateNode(n)
	})
	if err != nil {
		return NodeChange{}, err
	}
	if disabled {
		s.hub.Kick(id)
	}
	return s.answer(ctx, name, res, nil)
}

// Remove deletes a node.
func (s *Nodes) Remove(ctx context.Context, by Actor, name string) (NodeChange, error) {
	name, err := checkName(name)
	if err != nil {
		return NodeChange{}, err
	}
	var gone store.Node
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.rm", Target: name}, func(tx *store.Tx, c *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		gone = n
		c.Detail = map[string]any{"id": n.ID, "role": n.Role, "mesh_ip": n.MeshIP}
		return tx.DeleteNode(n.ID)
	})
	if err != nil {
		return NodeChange{}, err
	}
	s.hub.Kick(gone.ID)
	return NodeChange{Node: nodeInfo(gone, nil), Version: res.Version, Warnings: problemStrings(res.Warnings)}, nil
}

// Token issues a new join token, revoking unused ones.
func (s *Nodes) Token(ctx context.Context, by Actor, name string) (NodeChange, error) {
	name, err := checkName(name)
	if err != nil {
		return NodeChange{}, err
	}
	var enroll *Enrollment
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.token", Target: name}, func(tx *store.Tx, c *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		switch {
		case n.Role == string(proto.RolePhone):
			return errorf(http.StatusBadRequest, proto.CodeBadRequest,
				"node %q is a phone; a phone has no agent and no token: make its profile (`meshctl node phone %s`, or its page in the web UI)", n.Name, n.Name)
		case n.Disabled:
			return errorf(http.StatusConflict, proto.CodeConflict, "node %q is disabled; enable it first", n.Name)
		}
		token, exp, err := tx.IssueToken(n.ID, store.TokenTTL)
		if err != nil {
			return err
		}
		enroll = s.enrollment(token, exp.Unix())
		c.Detail = map[string]any{"rekey": n.Enrolled, "expires_at": exp.Unix()}
		return nil
	})
	if err != nil {
		return NodeChange{}, err
	}
	return s.answer(ctx, name, res, enroll)
}

// answer is a NodeChange with the node as it is now.
func (s *Nodes) answer(ctx context.Context, name string, res Result, enroll *Enrollment) (NodeChange, error) {
	var info NodeInfo
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		info = nodeInfo(n, nodes)
		return withHandshake(tx, &info)
	})
	if err != nil {
		return NodeChange{}, err
	}
	return NodeChange{Node: info, Version: res.Version, Enrollment: enroll, Warnings: problemStrings(res.Warnings)}, nil
}

// newNode validates an add request into a store row (mesh IP and MTU still
// to be filled in from the database).
func newNode(req NodeAddRequest) (store.Node, netip.Addr, error) {
	var (
		n    store.Node
		ip   netip.Addr
		errs []error
		err  error
	)
	keep := func(e error) {
		if e != nil {
			errs = append(errs, e)
		}
	}
	n.Name, err = validate.Name("name", req.Name)
	keep(err)
	n.Role, err = validate.Enum("role", req.Role, string(proto.RoleServer), string(proto.RoleSpoke), string(proto.RolePhone))
	keep(err)
	if req.IP != "" {
		ip, err = validate.IPv4("ip", req.IP)
		keep(err)
	}
	if req.Endpoint != "" {
		ap, err := validate.Endpoint("endpoint", req.Endpoint)
		keep(err)
		n.Endpoint = ap.String()
	}
	if n.Role != string(proto.RoleServer) && (req.Endpoint != "" || req.CanExit) {
		keep(errors.New("endpoint and can_exit apply to servers only; spokes dial out"))
	}
	n.ListenPort = proto.DefaultListenPort
	if req.ListenPort != 0 {
		n.ListenPort, err = validate.Port("listen_port", req.ListenPort)
		keep(err)
	}
	n.Keepalive = proto.DefaultKeepalive
	if req.Keepalive != 0 {
		n.Keepalive, err = validate.IntRange("keepalive", req.Keepalive, 1, 3600)
		keep(err)
	}
	if req.MTU != 0 {
		n.MTU, err = validate.IntRange("mtu", req.MTU, proto.MinMTU, proto.MaxMTU)
		keep(err)
	}
	if req.Hub != "" {
		_, err = validate.Name("hub", req.Hub)
		keep(err)
	}
	n.CanExit = req.CanExit
	if err := errors.Join(errs...); err != nil {
		return store.Node{}, netip.Addr{}, invalid(err)
	}
	return n, ip, nil
}

func defaultMTU(tx *store.Tx) (int, error) {
	s, err := tx.Setting(store.SettingDefaultMTU)
	if err != nil {
		return 0, err
	}
	mtu, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("setting %s = %q is not a number: %w", store.SettingDefaultMTU, s, err)
	}
	return mtu, nil
}

func validateSet(req NodeSetRequest) error {
	var errs []error
	check := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if req.Endpoint != nil && *req.Endpoint != "" {
		_, err := validate.Endpoint("endpoint", *req.Endpoint)
		check(err)
	}
	if req.ListenPort != nil {
		_, err := validate.Port("listen_port", *req.ListenPort)
		check(err)
	}
	if req.Hub != nil && *req.Hub != "" {
		_, err := validate.Name("hub", *req.Hub)
		check(err)
	}
	if req.Exit != nil && *req.Exit != "" {
		_, err := validate.Name("exit", *req.Exit)
		check(err)
	}
	if req.Keepalive != nil {
		_, err := validate.IntRange("keepalive", *req.Keepalive, 0, 3600)
		check(err)
	}
	if req.MTU != nil {
		_, err := validate.IntRange("mtu", *req.MTU, proto.MinMTU, proto.MaxMTU)
		check(err)
	}
	if err := errors.Join(errs...); err != nil {
		return invalid(err)
	}
	return nil
}

func (s *Nodes) enrollment(token string, expires int64) *Enrollment {
	return &Enrollment{
		Token:     token,
		ExpiresAt: expires,
		Install:   fmt.Sprintf("curl -fsSL %s%s | sudo sh -s -- --token %s", s.serverURL, proto.PathInstall, token),
		Join:      fmt.Sprintf("meshctl agent join --server %s --token %s", s.serverURL, token),
	}
}

func checkName(name string) (string, error) {
	name, err := validate.Name("node name", name)
	if err != nil {
		return "", invalid(err)
	}
	return name, nil
}

// nodeByName reads a node, turning "not found" into a 404 that names what
// was being looked up.
func nodeByName(tx *store.Tx, name, what string) (store.Node, error) {
	n, err := tx.NodeByName(name)
	if errors.Is(err, store.ErrNotFound) {
		return store.Node{}, errorf(http.StatusNotFound, proto.CodeNotFound, "%s %q does not exist; the node list shows every node", what, name)
	}
	return n, err
}

// refID resolves an optional reference by name; "" clears it.
func refID(tx *store.Tx, name, what string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	n, err := nodeByName(tx, name, what)
	return n.ID, err
}

func nodeInfo(n store.Node, all []store.Node) NodeInfo {
	nameOf := func(id int64) string {
		for _, o := range all {
			if o.ID == id {
				return o.Name
			}
		}
		return ""
	}
	info := NodeInfo{
		ID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, Endpoint: n.Endpoint, ListenPort: n.ListenPort,
		Hub: nameOf(n.PrimaryHubID), CanExit: n.CanExit, Exit: nameOf(n.ExitVia), Keepalive: n.Keepalive, MTU: n.MTU,
		Enrolled: n.Enrolled, Disabled: n.Disabled, OS: n.OS, Arch: n.Arch, Hostname: n.Hostname,
		WGBackend: n.WGBackend, AgentVersion: n.AgentVersion, ConfigVersion: n.ConfigVersion,
		AppliedVersion: n.AppliedVersion, InSync: n.ConfigHash != "" && n.AppliedHash == n.ConfigHash, LastSeen: n.LastSeen,
	}
	// The heartbeat's display-only part; written by status(), so it parses.
	var d statusDetail
	if json.Unmarshal([]byte(n.StatusJSON), &d) == nil {
		info.Degraded, info.Errors, info.Rollback, info.DNS = d.Degraded, d.Errors, d.Rollback, d.DNS
	}
	return info
}

func problemStrings(ps []planner.Problem) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.String())
	}
	return out
}
