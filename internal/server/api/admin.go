package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

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

// Nodes is the node administration service of one mesh, or of the sole
// mesh when it names none (amendment A49).
type Nodes struct {
	db        DB
	rec       *Reconciler
	hub       *Hub
	serverURL string
	// now is the clock the policies' schedules are judged on.
	now func() time.Time
	// mesh is the mesh the service works on; 0 is the sole mesh.
	mesh int64
}

// NewNodes wires the service from the API's dependencies. It works on the
// sole mesh; In picks another.
func NewNodes(d Deps) *Nodes {
	now := d.Now
	if now == nil {
		now = time.Now
	}
	return &Nodes{db: d.DB, rec: d.Reconciler, hub: d.Hub, serverURL: d.ServerURL, now: now}
}

// In is the service for one mesh.
func (s *Nodes) In(mesh int64) *Nodes {
	c := *s
	c.mesh = mesh
	return &c
}

// SoleMesh is the mesh a request that names none means: the only one the
// control plane has (amendment A49).
func SoleMesh(tx *store.Tx) (store.Mesh, error) {
	meshes, err := tx.Meshes()
	if err != nil {
		return store.Mesh{}, err
	}
	switch len(meshes) {
	case 1:
		return meshes[0], nil
	case 0:
		return store.Mesh{}, errorf(http.StatusNotFound, proto.CodeNotFound,
			"this control plane has no mesh: create one with `meshctl mesh add <name>`")
	}
	names := make([]string, len(meshes))
	for i, m := range meshes {
		names[i] = m.Name
	}
	return store.Mesh{}, errorf(http.StatusBadRequest, proto.CodeBadRequest,
		"this control plane has %d meshes (%s): name one with --mesh", len(meshes), strings.Join(names, ", "))
}

// meshID is the service's mesh: its own, else the sole mesh.
func (s *Nodes) meshID(tx *store.Tx) (int64, error) {
	if s.mesh != 0 {
		return s.mesh, nil
	}
	m, err := SoleMesh(tx)
	return m.ID, err
}

// tx runs fn in a transaction scoped to the service's mesh.
func (s *Nodes) tx(ctx context.Context, fn func(*store.MeshTx) error) error {
	return s.db.Tx(ctx, func(raw *store.Tx) error {
		mesh, err := s.meshID(raw)
		if err != nil {
			return err
		}
		return fn(raw.In(mesh))
	})
}

// resolve is the service's mesh, reading the sole mesh when it names none.
func (s *Nodes) resolve(ctx context.Context) (int64, error) {
	if s.mesh != 0 {
		return s.mesh, nil
	}
	var mesh int64
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		var err error
		mesh, err = s.meshID(tx)
		return err
	})
	return mesh, err
}

// apply is the reconciler's Apply in the service's mesh.
func (s *Nodes) apply(ctx context.Context, c Change, mutate func(*store.MeshTx, *Change) error) (Result, error) {
	var err error
	if c.Mesh, err = s.resolve(ctx); err != nil {
		return Result{}, err
	}
	return s.rec.Apply(ctx, c, mutate)
}

// Tx runs fn in a transaction scoped to the service's mesh: the web UI's
// reads go through it.
func (s *Nodes) Tx(ctx context.Context, fn func(*store.MeshTx) error) error { return s.tx(ctx, fn) }

// Apply is the reconciler's Apply in the service's mesh.
func (s *Nodes) Apply(ctx context.Context, c Change, mutate func(*store.MeshTx, *Change) error) (Result, error) {
	return s.apply(ctx, c, mutate)
}

// Rollback is the reconciler's Rollback in the service's mesh.
func (s *Nodes) Rollback(ctx context.Context, c Change, n int64) (Result, error) {
	var err error
	if c.Mesh, err = s.resolve(ctx); err != nil {
		return Result{}, err
	}
	return s.rec.Rollback(ctx, c, n)
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
		{Route{http.MethodPost, one + "/phone", "socket"}, a.adminPhone},
		{Route{http.MethodGet, AdminSegmentation, "socket"}, a.adminSegment},
		{Route{http.MethodPut, AdminSegmentation, "socket"}, a.adminSetSegment},
		{Route{http.MethodGet, AdminSegmentation + "/check", "socket"}, a.adminReach},
		{Route{http.MethodGet, AdminIPv6, "socket"}, a.adminIPv6},
		{Route{http.MethodPut, AdminIPv6, "socket"}, a.adminSetIPv6},
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

func (a *api) readJSON(w http.ResponseWriter, r *http.Request, v any) error {
	body, err := readBody(w, r)
	if err != nil {
		return err
	}
	return decodeJSON(body, v)
}

// containerNetwork is n's container network for Show, nil without one.
func containerNetwork(in planner.Input, n store.Node) *ContainerNetwork {
	k := planner.ContainerNetwork(in.MeshCIDR, in.MeshCIDR6, plannerNode(n))
	if k == nil {
		return nil
	}
	out := &ContainerNetwork{
		Name: proto.MeshNetwork, BridgeName: proto.ContainerBridge, Mode: k.Mode, Subnet: k.Subnet, Network: k.Network,
		Range: k.Range, Bridge: k.Bridge, Gateway: k.Gateway, Zone: in.ZoneName(), Resolvers: planner.ServerIPs(in),
		Subnet6: k.Subnet6, Network6: k.Network6, Range6: k.Range6, Bridge6: k.Bridge6, Gateway6: k.Gateway6,
	}
	if k.Mode == proto.ModeContainer && n.Role == string(proto.RoleServer) {
		// The endpoint names the host's port; none known yet leaves it out.
		ep, err := netip.ParseAddrPort(n.Endpoint)
		if err != nil {
			ep, err = netip.ParseAddrPort(n.Endpoint6)
		}
		if err == nil {
			out.PublishPort = int(ep.Port())
		}
		out.ListenPort = n.ListenPort
	}
	return out
}

// List returns every node and the plan's problems.
func (s *Nodes) List(ctx context.Context) (NodeList, error) {
	var out NodeList
	err := s.tx(ctx, func(tx *store.MeshTx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		groups, err := tx.NodeGroups()
		if err != nil {
			return err
		}
		out.Nodes = make([]NodeInfo, 0, len(nodes))
		for _, n := range nodes {
			info := nodeInfo(n, nodes)
			info.Groups = groups[n.ID]
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
	err = s.tx(ctx, func(tx *store.MeshTx) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		out.Node = nodeInfo(n, nodes)
		out.Containers = containerNetwork(in, n)
		if err := withGroups(tx, &out.Node); err != nil {
			return err
		}
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
	n, fixedIP, fixedIP6, err := newNode(req)
	if err != nil {
		return NodeChange{}, err
	}
	groups, err := checkGroups(req.Groups)
	if err != nil {
		return NodeChange{}, err
	}
	var enroll *Enrollment
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.add", Target: n.Name}, func(tx *store.MeshTx, c *Change) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		used := usedAddrs(nodes, 0)
		block := planner.RangeFor(proto.Role(n.Role), req.AlwaysOn)
		switch {
		case n.Family == proto.FamilyIPv6:
		case fixedIP.IsValid():
			if err := planner.CheckRequested(in.MeshCIDR, fixedIP, used); err != nil {
				return invalid(err)
			}
			n.MeshIP = fixedIP.String()
		default:
			ip, err := planner.Allocate(in.MeshCIDR, block, used)
			if err != nil {
				return errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
			}
			n.MeshIP = ip.String()
		}
		if n.Family != proto.FamilyIPv4 {
			if err := addIPv6(in, &n, fixedIP6, block, used); err != nil {
				return err
			}
		}
		if n.Mode == proto.ModeContainer || req.Containers {
			if ip, err := netip.ParseAddr(n.MeshIP); err == nil {
				used[ip] = true
			}
			if n.ContainerSubnet, err = containerSubnet(in.MeshCIDR, nodes, used); err != nil {
				return err
			}
		}
		if req.Containers6 {
			if n.ContainerSubnet6, err = containerSubnet6(in.MeshCIDR6, nodes); err != nil {
				return err
			}
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
		if len(groups) > 0 {
			if err := setGroups(tx, n.ID, groups); err != nil {
				return err
			}
		}
		detail := map[string]any{"role": n.Role, "mesh_ip": n.MeshIP, "endpoint": n.Endpoint, "hub": req.Hub, "can_exit": n.CanExit}
		if len(groups) > 0 {
			detail["groups"] = groups
		}
		if n.Mode != proto.ModeHost {
			detail["mode"] = n.Mode
		}
		if n.ContainerSubnet != "" {
			detail["container_subnet"] = n.ContainerSubnet
		}
		for k, v := range map[string]string{"mesh_ip6": n.MeshIP6, "endpoint6": n.Endpoint6, "container_subnet6": n.ContainerSubnet6} {
			if v != "" {
				detail[k] = v
			}
		}
		if n.Family != proto.FamilyDual {
			detail["ip_family"] = n.Family
		}
		c.Detail = detail
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
	var notes []string
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.set", Target: name}, func(tx *store.MeshTx, c *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		detail := map[string]any{}
		if req.Family != nil {
			if err := setFamily(tx, &n, *req.Family); err != nil {
				return err
			}
			detail["ip_family"], detail["mesh_ip"], detail["mesh_ip6"] = n.Family, n.MeshIP, n.MeshIP6
		}
		if req.MeshIP != nil {
			if notes, err = moveMeshIP(tx, &n, *req.MeshIP); err != nil {
				return err
			}
			detail["mesh_ip"] = n.MeshIP
		}
		if req.MeshIP6 != nil {
			more, err := moveMeshIP6(tx, &n, *req.MeshIP6)
			if err != nil {
				return err
			}
			notes = append(notes, more...)
			detail["mesh_ip6"] = n.MeshIP6
		}
		if req.Endpoint6 != nil {
			n.Endpoint6 = *req.Endpoint6
			if n.Endpoint6 != "" {
				ap, _ := validate.Endpoint6("endpoint6", n.Endpoint6) // validated in validateSet
				n.Endpoint6 = ap.String()
			}
			detail["endpoint6"] = n.Endpoint6
		}
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
		if req.Containers != nil {
			if err := setContainers(tx, &n, *req.Containers); err != nil {
				return err
			}
			detail["container_subnet"] = n.ContainerSubnet
		}
		if req.Containers6 != nil {
			if err := setContainers6(tx, &n, *req.Containers6); err != nil {
				return err
			}
			detail["container_subnet6"] = n.ContainerSubnet6
		}
		if req.Groups != nil {
			groups, _ := checkGroups(*req.Groups) // validated in validateSet
			if err := setGroups(tx, n.ID, groups); err != nil {
				return err
			}
			detail["groups"] = groups
		}
		c.Detail = detail
		return tx.UpdateNode(n)
	})
	if err != nil {
		return NodeChange{}, err
	}
	ch, err := s.answer(ctx, name, res, nil)
	ch.Notes = notes
	return ch, err
}

// usedAddrs is every mesh address of nodes, of both families, but the
// node except's.
func usedAddrs(nodes []store.Node, except int64) map[netip.Addr]bool {
	used := map[netip.Addr]bool{}
	for _, o := range nodes {
		if o.ID == except {
			continue
		}
		for _, s := range []string{o.MeshIP, o.MeshIP6} {
			if ip, err := netip.ParseAddr(s); err == nil {
				used[ip] = true
			}
		}
	}
	return used
}

// addIPv6 gives n an IPv6 mesh address (M8): fixed when the operator chose
// one, else the lowest free one of block; none while the mesh has no IPv6
// prefix, unless n is IPv6 only, which needs one.
func addIPv6(in planner.Input, n *store.Node, fixed netip.Addr, block planner.Range, used map[netip.Addr]bool) error {
	switch {
	case !in.MeshCIDR6.IsValid() && (fixed.IsValid() || n.Family == proto.FamilyIPv6):
		return errorf(http.StatusConflict, proto.CodeConflict,
			"the mesh has no IPv6 prefix yet, so %s cannot have an IPv6 address; set one with `meshctl ipv6 auto` (or a prefix) first", n.Name)
	case !in.MeshCIDR6.IsValid():
		return nil
	case fixed.IsValid():
		if err := planner.CheckRequested6(in.MeshCIDR6, fixed, used); err != nil {
			return invalid(err)
		}
		n.MeshIP6 = fixed.String()
	default:
		ip, err := planner.Allocate6(in.MeshCIDR6, block, used)
		if err != nil {
			return errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
		}
		n.MeshIP6 = ip.String()
	}
	used[netip.MustParseAddr(n.MeshIP6)] = true
	return nil
}

// blockOf is the address block n's existing address is in, so the family
// it gains is allocated alike (an always-on spoke stays always-on).
func blockOf(in planner.Input, n store.Node) planner.Range {
	if ip, err := netip.ParseAddr(n.MeshIP); err == nil && in.MeshCIDR.Contains(ip) {
		a, m := ip.As4(), in.MeshCIDR.Addr().As4()
		if r := (int(a[2]) - int(m[2])) + (int(a[1])-int(m[1]))*256; r >= 0 && r <= int(planner.RangePhones) {
			return planner.Range(r)
		}
	}
	if ip, err := netip.ParseAddr(n.MeshIP6); err == nil {
		b := ip.As16()
		if r := int(b[12])<<8 | int(b[13]); r <= int(planner.RangePhones) {
			return planner.Range(r)
		}
	}
	return planner.RangeFor(proto.Role(n.Role), false)
}

// setFamily changes n's address families (M8): a family it gains gets an
// address allocated (IPv6 only once the mesh has a prefix), one it loses
// takes its address away -- and, for IPv6, its containers' /112.
func setFamily(tx *store.MeshTx, n *store.Node, family string) error {
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return err
	}
	used, block := usedAddrs(nodes, n.ID), blockOf(in, *n)
	switch family {
	case proto.FamilyIPv4:
		n.MeshIP6, n.ContainerSubnet6 = "", ""
	case proto.FamilyIPv6:
		if n.ContainerSubnet != "" {
			return errorf(http.StatusConflict, proto.CodeConflict,
				"%s has a container network, which needs IPv4; take its containers off the mesh first (`meshctl node set %s --containers=false`)", n.Name, n.Name)
		}
		n.MeshIP = ""
	}
	n.Family = family
	if family != proto.FamilyIPv6 && n.MeshIP == "" {
		ip, err := planner.Allocate(in.MeshCIDR, block, used)
		if err != nil {
			return errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
		}
		n.MeshIP = ip.String()
	}
	if family != proto.FamilyIPv4 && n.MeshIP6 == "" {
		return addIPv6(in, n, netip.Addr{}, block, used)
	}
	return nil
}

// moveMeshIP6 gives n the IPv6 mesh address ip (M8), checked as a fixed
// one at add is; the same address changes nothing.
func moveMeshIP6(tx *store.MeshTx, n *store.Node, ip string) ([]string, error) {
	addr, _ := validate.IPv6("mesh_ip6", ip) // validated in validateSet
	old := n.MeshIP6
	if addr.String() == old {
		return nil, nil
	}
	if n.Family == proto.FamilyIPv4 {
		return nil, errorf(http.StatusBadRequest, proto.CodeBadRequest, "%s is IPv4 only; make it dual first (`meshctl node set %s --family dual`)", n.Name, n.Name)
	}
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return nil, err
	}
	if !in.MeshCIDR6.IsValid() {
		return nil, errorf(http.StatusConflict, proto.CodeConflict, "the mesh has no IPv6 prefix yet; set one with `meshctl ipv6 auto` (or a prefix) first")
	}
	if err := planner.CheckRequested6(in.MeshCIDR6, addr, usedAddrs(nodes, n.ID)); err != nil {
		return nil, invalid(err)
	}
	n.MeshIP6 = addr.String()
	if old == "" {
		return nil, nil
	}
	notes := []string{fmt.Sprintf("%s moves from %s to %s; open IPv6 connections to the old address drop", n.Name, old, n.MeshIP6)}
	if n.Role == string(proto.RolePhone) {
		notes = append(notes, fmt.Sprintf("the phone's profile still has %s: make a new one (`meshctl node phone %s`)", old, n.Name))
	}
	return notes, nil
}

// moveMeshIP gives n the mesh IP ip (amendment A41), checked as a fixed IP
// at add is: inside the mesh, not the network or broadcast address of its
// /24, not another node's; the plan check that follows refuses one inside a
// container subnet. The same address changes nothing. The notes are what
// the move does not reach.
func moveMeshIP(tx *store.MeshTx, n *store.Node, ip string) ([]string, error) {
	addr, _ := validate.IPv4("mesh_ip", ip) // validated in validateSet
	old := n.MeshIP
	if addr.String() == old {
		return nil, nil
	}
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return nil, err
	}
	used := map[netip.Addr]bool{}
	for _, o := range nodes {
		if a, err := netip.ParseAddr(o.MeshIP); err == nil && o.ID != n.ID {
			used[a] = true
		}
	}
	if err := planner.CheckRequested(in.MeshCIDR, addr, used); err != nil {
		return nil, invalid(err)
	}
	n.MeshIP = addr.String()
	notes := []string{fmt.Sprintf("%s moves from %s to %s; open connections to the old address drop", n.Name, old, n.MeshIP)}
	switch n.Role {
	case string(proto.RoleServer):
		notes = append(notes, fmt.Sprintf("anything set up with %s on %s keeps it -- Swarm's advertise address above all: rejoin Swarm there with --advertise-addr %s", old, n.Name, n.MeshIP))
	case string(proto.RolePhone):
		notes = append(notes, fmt.Sprintf("the phone's profile still has %s: make a new one (`meshctl node phone %s`)", old, n.Name))
	}
	seg, err := tx.Segmentation()
	if err != nil {
		return nil, err
	}
	oldIP, _ := netip.ParseAddr(old) // stored validated
	for _, l := range segLiterals(seg) {
		if r, err := validate.AddrEntry("entry", l.entry); err == nil && inRange(r, oldIP) && !inRange(r, addr) {
			notes = append(notes, fmt.Sprintf("%s names %s by address, which held the old one; node:%s follows the node", l.where, l.entry, n.Name))
		}
	}
	return notes, nil
}

// segLiteral is an address written into the segmentation, and where.
type segLiteral struct{ where, entry string }

// segLiterals are the literal addresses of address objects, zone members
// and policies' From and To.
func segLiterals(s store.Segmentation) []segLiteral {
	var out []segLiteral
	add := func(where string, vs []string) {
		for _, v := range vs {
			if !strings.Contains(v, ":") {
				out = append(out, segLiteral{where, v})
			}
		}
	}
	for _, o := range s.Addresses {
		add(fmt.Sprintf("address object %q", o.Name), o.Entries)
	}
	for _, z := range s.Zones {
		add(fmt.Sprintf("zone %q", z.Name), z.Members)
	}
	for _, p := range s.Policies {
		add(fmt.Sprintf("policy %q", p.Name), append(slices.Clone(p.From), p.To...))
	}
	return out
}

func inRange(r validate.AddrRange, a netip.Addr) bool { return !a.Less(r.First) && !r.Last.Less(a) }

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
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: action, Target: name}, func(tx *store.MeshTx, _ *Change) error {
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
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.rm", Target: name}, func(tx *store.MeshTx, c *Change) error {
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
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "node.token", Target: name}, func(tx *store.MeshTx, c *Change) error {
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
	err := s.tx(ctx, func(tx *store.MeshTx) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		info = nodeInfo(n, nodes)
		if err := withGroups(tx, &info); err != nil {
			return err
		}
		return withHandshake(tx, &info)
	})
	if err != nil {
		return NodeChange{}, err
	}
	return NodeChange{Node: info, Version: res.Version, Enrollment: enroll, Warnings: problemStrings(res.Warnings)}, nil
}

// newNode validates an add request into a store row (mesh IP and MTU still
// to be filled in from the database).
func newNode(req NodeAddRequest) (store.Node, netip.Addr, netip.Addr, error) {
	var (
		n       store.Node
		ip, ip6 netip.Addr
		errs    []error
		err     error
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
	if req.Endpoint6 != "" {
		ap, err := validate.Endpoint6("endpoint6", req.Endpoint6)
		keep(err)
		n.Endpoint6 = ap.String()
	}
	if n.Role != string(proto.RoleServer) && (req.Endpoint != "" || req.Endpoint6 != "" || req.CanExit) {
		keep(errors.New("endpoint, endpoint6 and can_exit apply to servers only; spokes dial out"))
	}
	n.Family = proto.FamilyDual
	if req.Family != "" {
		n.Family, err = validate.Enum("ip_family", req.Family, proto.FamilyDual, proto.FamilyIPv4, proto.FamilyIPv6)
		keep(err)
	}
	if req.IP6 != "" {
		ip6, err = validate.IPv6("ip6", req.IP6)
		keep(err)
	}
	switch {
	case n.Family == proto.FamilyIPv4 && req.IP6 != "":
		keep(errors.New("an ipv4 node has no IPv6 mesh address; leave ip6 out, or make it dual"))
	case n.Family == proto.FamilyIPv6 && req.IP != "":
		keep(errors.New("an ipv6 node has no IPv4 mesh IP; leave ip out, or make it dual"))
	case req.Containers6 && n.Family == proto.FamilyIPv4:
		keep(errors.New("containers6 needs IPv6, and an ipv4 node has none"))
	case req.Containers6 && n.Family == proto.FamilyIPv6:
		keep(errors.New("an ipv6 node has no container network: containers need IPv4, and their IPv6 comes beside it"))
	case req.Containers6 && !req.Containers && req.Mode != proto.ModeContainer:
		keep(errors.New("containers6 puts the containers' IPv6 beside their IPv4: give the node a container network too (containers, or container mode)"))
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
	n.Mode = proto.ModeHost
	if req.Mode != "" {
		n.Mode, err = validate.Enum("mode", req.Mode, proto.ModeHost, proto.ModeContainer)
		keep(err)
	}
	container := n.Mode == proto.ModeContainer
	switch {
	case n.Role == string(proto.RolePhone) && (container || req.Containers):
		keep(errors.New("a phone has no agent, so it has no container network; mode and containers apply to servers and spokes"))
	case container && n.Role == string(proto.RoleServer) && req.Endpoint == "" && req.Endpoint6 == "":
		keep(errors.New("a container-mode server needs an endpoint, <public IPv4>:<port the host publishes>: inside its container the agent cannot see the host's address"))
	case container && req.CanExit:
		keep(errors.New("a container-mode server cannot be an exit: its host is not on the mesh"))
	}
	if err := errors.Join(errs...); err != nil {
		return store.Node{}, netip.Addr{}, netip.Addr{}, invalid(err)
	}
	return n, ip, ip6, nil
}

// containerSubnet allocates the lowest free container /24 (amendment A37);
// ips are the mesh IPs in use, the new node's included.
func containerSubnet(mesh netip.Prefix, nodes []store.Node, ips map[netip.Addr]bool) (string, error) {
	taken := map[netip.Prefix]bool{}
	for _, o := range nodes {
		if k, err := netip.ParsePrefix(o.ContainerSubnet); err == nil {
			taken[k] = true
		}
	}
	k, err := planner.AllocateContainers(mesh, taken, ips)
	if err != nil {
		return "", errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
	}
	return k.String(), nil
}

// containerSubnet6 allocates the lowest free container /112 of mesh6 (M8).
func containerSubnet6(mesh6 netip.Prefix, nodes []store.Node) (string, error) {
	if !mesh6.IsValid() {
		return "", errorf(http.StatusConflict, proto.CodeConflict,
			"the mesh has no IPv6 prefix yet, so containers have no IPv6 to get; set one with `meshctl ipv6 auto` (or a prefix) first")
	}
	taken := map[netip.Prefix]bool{}
	for _, o := range nodes {
		if k, err := netip.ParsePrefix(o.ContainerSubnet6); err == nil {
			taken[k] = true
		}
	}
	k, err := planner.AllocateContainers6(mesh6, taken)
	if err != nil {
		return "", errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
	}
	return k.String(), nil
}

// setContainers6 puts n's containers' IPv6 on the mesh or takes it off
// (M8): it comes beside their IPv4 network, on a node with IPv6.
func setContainers6(tx *store.MeshTx, n *store.Node, on bool) error {
	switch {
	case !on:
		n.ContainerSubnet6 = ""
		return nil
	case n.ContainerSubnet6 != "":
		return nil
	case n.ContainerSubnet == "":
		return errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"%s has no container network; their IPv6 comes beside their IPv4, so give it one first (`meshctl node set %s --containers`)", n.Name, n.Name)
	case n.MeshIP6 == "":
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "%s has no IPv6 mesh address (ip_family %s), so its containers cannot have IPv6", n.Name, n.Family)
	}
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return err
	}
	n.ContainerSubnet6, err = containerSubnet6(in.MeshCIDR6, nodes)
	return err
}

// setContainers gives a host-mode node a container subnet (on) or takes it
// away; a container-mode node keeps its own, and a phone has none.
func setContainers(tx *store.MeshTx, n *store.Node, on bool) error {
	switch {
	case n.Role == string(proto.RolePhone):
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "%s is a phone; a phone has no agent, so it has no container network", n.Name)
	case n.Mode == proto.ModeContainer && !on:
		return errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"%s is in container mode, where its containers are always on the mesh; to change its mode, remove the node and add it again", n.Name)
	case !on:
		n.ContainerSubnet, n.ContainerSubnet6 = "", "" // the IPv6 side comes beside the IPv4 one
		return nil
	case n.ContainerSubnet != "":
		return nil
	}
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return err
	}
	ips := map[netip.Addr]bool{}
	for _, o := range nodes {
		if ip, err := netip.ParseAddr(o.MeshIP); err == nil {
			ips[ip] = true
		}
	}
	n.ContainerSubnet, err = containerSubnet(in.MeshCIDR, nodes, ips)
	return err
}

func defaultMTU(tx *store.MeshTx) (int, error) {
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
	if req.MeshIP != nil {
		_, err := validate.IPv4("mesh_ip", *req.MeshIP)
		check(err)
	}
	if req.MeshIP6 != nil {
		_, err := validate.IPv6("mesh_ip6", *req.MeshIP6)
		check(err)
	}
	if req.Endpoint6 != nil && *req.Endpoint6 != "" {
		_, err := validate.Endpoint6("endpoint6", *req.Endpoint6)
		check(err)
	}
	if req.Family != nil {
		_, err := validate.Enum("ip_family", *req.Family, proto.FamilyDual, proto.FamilyIPv4, proto.FamilyIPv6)
		check(err)
	}
	if req.Groups != nil {
		if _, err := checkGroups(*req.Groups); err != nil {
			return err
		}
	}
	if err := errors.Join(errs...); err != nil {
		return invalid(err)
	}
	return nil
}

// checkGroups validates group names: each a Name, once; sorted.
func checkGroups(in []string) ([]string, error) {
	seen := map[string]bool{}
	for i, g := range in {
		if _, err := validate.Name(fmt.Sprintf("groups[%d]", i), g); err != nil {
			return nil, invalid(err)
		}
		if seen[g] {
			return nil, invalid(&validate.Error{Field: "groups", Value: g, Reason: "is listed twice"})
		}
		seen[g] = true
	}
	return slices.Sorted(maps.Keys(seen)), nil
}

// setGroups puts node id in exactly groups and checks the segmentation
// still keeps its limits.
func setGroups(tx *store.MeshTx, id int64, groups []string) error {
	if err := tx.SetNodeGroups(id, groups); err != nil {
		return err
	}
	seg, err := tx.Segmentation()
	if err != nil {
		return err
	}
	if err := seg.Validate(); err != nil {
		return invalid(err)
	}
	return nil
}

// withGroups fills in the node's groups.
func withGroups(tx *store.MeshTx, info *NodeInfo) error {
	groups, err := tx.NodeGroups()
	info.Groups = groups[info.ID]
	return err
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
func nodeByName(tx *store.MeshTx, name, what string) (store.Node, error) {
	n, err := tx.NodeByName(name)
	if errors.Is(err, store.ErrNotFound) {
		return store.Node{}, errorf(http.StatusNotFound, proto.CodeNotFound, "%s %q does not exist; the node list shows every node", what, name)
	}
	return n, err
}

// refID resolves an optional reference by name; "" clears it.
func refID(tx *store.MeshTx, name, what string) (int64, error) {
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
		Mode: n.Mode, ContainerSubnet: n.ContainerSubnet, DownSince: n.DownSince,
		MeshIP6: n.MeshIP6, Family: n.Family, Endpoint6: n.Endpoint6, ContainerSubnet6: n.ContainerSubnet6,
	}
	if n.HostFamilies != "" {
		info.HostFamilies = strings.Split(n.HostFamilies, ",")
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
