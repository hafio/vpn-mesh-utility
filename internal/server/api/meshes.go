package api

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The control plane's meshes (amendment A49). An admin adds, renames and
// removes them; every node, setting, record, version and policy belongs to
// one. Meshes of one control plane share no address and no DNS domain,
// since agents of two meshes may run on one host: every write is checked
// against the other meshes (checkApart).

// AdminMeshes is the admin API's meshes.
const AdminMeshes = "/admin/v1/meshes"

// New meshes' settings: the first mesh's, but for the network and the zone.
const (
	defaultUpstream = "1.1.1.1,9.9.9.9"
	defaultMeshMTU  = "1420"
)

// MeshInfo is one mesh as the admin API shows it.
type MeshInfo struct {
	Name string `json:"name"`
	// UUID names the mesh to its agents; Slot (0..63) picks the names its
	// agents use on a host. Neither ever changes.
	UUID       string   `json:"uuid"`
	Slot       int      `json:"slot"`
	CreatedAt  int64    `json:"created_at"`
	Nodes      int      `json:"nodes"`
	MeshCIDR   string   `json:"mesh_cidr"`
	MeshCIDR6  string   `json:"mesh_cidr6,omitempty"`
	Zone       string   `json:"dns_zone"`
	Domains    []string `json:"dns_extra_domains,omitempty"`
	Upstream   []string `json:"upstream_dns"`
	DefaultMTU int      `json:"default_mtu"`
	DefaultHub string   `json:"default_hub,omitempty"`
}

// MeshList is GET /admin/v1/meshes.
type MeshList struct {
	Meshes []MeshInfo `json:"meshes"`
}

// MeshAddRequest is POST /admin/v1/meshes. An empty field takes its
// default: the first free 10.x.0.0/16, the name as the DNS zone, and IPv6
// auto (a random unique local /64; IPv6Off for none).
type MeshAddRequest struct {
	Name     string `json:"name"`
	MeshCIDR string `json:"mesh_cidr,omitempty"`
	Zone     string `json:"dns_zone,omitempty"`
	IPv6     string `json:"ipv6,omitempty"`
}

// SettingsRequest changes a mesh's settings (the web UI's Settings page,
// `meshctl mesh set`); only non-nil fields change. The IPv4 network
// changes only while the mesh has no nodes; DefaultHub names a server, ""
// for none. The IPv6 prefix has its own request (IPv6Request).
type SettingsRequest struct {
	MeshCIDR   *string   `json:"mesh_cidr,omitempty"`
	Zone       *string   `json:"dns_zone,omitempty"`
	Domains    *[]string `json:"dns_extra_domains,omitempty"`
	Upstream   *[]string `json:"upstream_dns,omitempty"`
	DefaultMTU *int      `json:"default_mtu,omitempty"`
	DefaultHub *string   `json:"default_hub,omitempty"`
}

// MeshSetRequest is PATCH /admin/v1/meshes/{name}: a new name, the
// settings, or both.
type MeshSetRequest struct {
	Name *string `json:"name,omitempty"`
	SettingsRequest
}

// MeshUpdate is the answer to a mesh write.
type MeshUpdate struct {
	Mesh MeshInfo `json:"mesh"`
	// Version is the config version the change created; 0 when no node's
	// computed config changed.
	Version  int64    `json:"version"`
	Warnings []string `json:"warnings,omitempty"`
}

// PickMesh is the mesh a request names, or the sole mesh when it names
// none.
func PickMesh(tx *store.Tx, name string) (store.Mesh, error) {
	if name == "" {
		return SoleMesh(tx)
	}
	m, err := tx.MeshByName(name)
	if errors.Is(err, store.ErrNotFound) {
		return store.Mesh{}, errorf(http.StatusNotFound, proto.CodeNotFound,
			"this control plane has no mesh %q: `meshctl mesh list` shows its meshes", name)
	}
	return m, err
}

// Named is the service for the mesh called name, looked up in each
// transaction; "" is the sole mesh.
func (s *Nodes) Named(name string) *Nodes {
	c := *s
	c.mesh, c.meshName = 0, name
	return &c
}

// Meshes lists the control plane's meshes.
func (s *Nodes) Meshes(ctx context.Context) (MeshList, error) {
	out := MeshList{Meshes: []MeshInfo{}}
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		meshes, err := tx.Meshes()
		if err != nil {
			return err
		}
		for _, m := range meshes {
			info, err := meshInfo(tx.In(m.ID), m)
			if err != nil {
				return err
			}
			out.Meshes = append(out.Meshes, info)
		}
		return nil
	})
	return out, err
}

// ShowMesh returns one mesh.
func (s *Nodes) ShowMesh(ctx context.Context, name string) (MeshInfo, error) {
	if _, err := validate.Name("mesh", name); err != nil {
		return MeshInfo{}, invalid(err)
	}
	var out MeshInfo
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		m, err := PickMesh(tx, name)
		if err != nil {
			return err
		}
		out, err = meshInfo(tx.In(m.ID), m)
		return err
	})
	return out, err
}

// AddMesh creates a mesh in the lowest free slot, with req's network, zone
// and IPv6 prefix or their defaults, the first mesh's other settings and the
// allow-all policy. It has no nodes, so no config changes.
func (s *Nodes) AddMesh(ctx context.Context, by Actor, req MeshAddRequest) (MeshUpdate, error) {
	name, err := validate.Name("name", req.Name)
	if err != nil {
		return MeshUpdate{}, invalid(err)
	}
	zone := strings.TrimSpace(req.Zone)
	if zone == "" {
		zone = name
	}
	if _, err := validate.Zone("dns_zone", zone); err != nil {
		return MeshUpdate{}, invalid(err)
	}
	var cidr, cidr6 netip.Prefix
	if req.MeshCIDR != "" {
		if cidr, err = meshCIDR(req.MeshCIDR); err != nil {
			return MeshUpdate{}, invalid(err)
		}
	}
	v6 := strings.TrimSpace(req.IPv6)
	switch v6 {
	case "", IPv6Auto:
		if cidr6, err = ULAPrefix(rand.Read); err != nil {
			return MeshUpdate{}, err
		}
	case IPv6Off:
	default:
		if cidr6, err = validate.Prefix6("ipv6", v6); err == nil {
			err = planner.CheckMeshCIDR6(cidr6)
		}
		if err != nil {
			return MeshUpdate{}, invalid(err)
		}
	}
	var out MeshUpdate
	err = s.db.Tx(ctx, func(raw *store.Tx) error {
		meshes, err := raw.Meshes()
		if err != nil {
			return err
		}
		if slices.ContainsFunc(meshes, func(m store.Mesh) bool { return m.Name == name }) {
			return errorf(http.StatusConflict, proto.CodeConflict, "mesh %q exists already", name)
		}
		slot, err := freeSlot(meshes)
		if err != nil {
			return err
		}
		if !cidr.IsValid() {
			spaces, err := spacesOf(raw, meshes)
			if err != nil {
				return err
			}
			if cidr, err = freeCIDR(spaces); err != nil {
				return err
			}
		}
		m := store.Mesh{Name: name, Slot: slot}
		if err := raw.InsertMesh(&m); err != nil {
			return err
		}
		tx := raw.In(m.ID)
		values := map[string]string{store.SettingMeshCIDR: cidr.String(), store.SettingDNSZone: zone,
			store.SettingUpstreamDNS: defaultUpstream, store.SettingDefaultMTU: defaultMeshMTU, store.SettingDefaultHub: "",
			store.SettingMeshCIDR6: store.MeshCIDR6Off}
		if cidr6.IsValid() {
			values[store.SettingMeshCIDR6] = cidr6.String()
		}
		for _, k := range slices.Sorted(maps.Keys(values)) {
			if err := tx.SetSetting(k, values[k]); err != nil {
				return err
			}
		}
		if err := tx.SetSegmentation(store.Segmentation{Policies: []store.SegPolicy{store.DefaultPolicy()}}); err != nil {
			return err
		}
		if err := checkApart(tx); err != nil {
			return err
		}
		detail := map[string]any{"uuid": m.UUID, "slot": m.Slot, "mesh_cidr": values[store.SettingMeshCIDR],
			"mesh_cidr6": values[store.SettingMeshCIDR6], "dns_zone": zone}
		if err := tx.Audit(by.Name, by.UserID, "mesh.add", name, detail); err != nil {
			return err
		}
		out.Mesh, err = meshInfo(tx, m)
		return err
	})
	return out, err
}

// SetMesh renames a mesh, changes its settings, or both, as one change
// through the reconciler.
func (s *Nodes) SetMesh(ctx context.Context, by Actor, name string, req MeshSetRequest) (MeshUpdate, error) {
	if _, err := validate.Name("mesh", name); err != nil {
		return MeshUpdate{}, invalid(err)
	}
	if req.Name != nil {
		if _, err := validate.Name("name", *req.Name); err != nil {
			return MeshUpdate{}, invalid(err)
		}
	}
	if req.Name == nil && req.SettingsRequest == (SettingsRequest{}) {
		return MeshUpdate{}, errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"nothing to change: give a new name or a setting (`meshctl mesh set --help`)")
	}
	write, err := s.settings(req.SettingsRequest)
	if err != nil {
		return MeshUpdate{}, err
	}
	in := s.Named(name)
	res, err := in.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "mesh.set", Target: name},
		func(tx *store.MeshTx, c *Change) error {
			if err := write(tx, c); err != nil {
				return err
			}
			if req.Name == nil || *req.Name == name {
				return nil
			}
			if _, err := tx.MeshByName(*req.Name); err == nil {
				return errorf(http.StatusConflict, proto.CodeConflict, "mesh %q exists already", *req.Name)
			} else if !errors.Is(err, store.ErrNotFound) {
				return err
			}
			c.Detail = map[string]any{"settings": c.Detail, "name": *req.Name}
			return tx.RenameMesh(tx.MeshID(), *req.Name)
		})
	if err != nil {
		return MeshUpdate{}, err
	}
	out := MeshUpdate{Version: res.Version, Warnings: problemStrings(res.Warnings)}
	if req.Name != nil {
		name = *req.Name
	}
	out.Mesh, err = s.ShowMesh(ctx, name)
	return out, err
}

// SetSettings changes the service's mesh's settings through the
// reconciler: a new default hub or MTU changes configs.
func (s *Nodes) SetSettings(ctx context.Context, by Actor, req SettingsRequest) (MeshChange, error) {
	write, err := s.settings(req)
	if err != nil {
		return MeshChange{}, err
	}
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "settings.set"}, write)
	if err != nil {
		return MeshChange{}, err
	}
	return MeshChange{Version: res.Version, Warnings: problemStrings(res.Warnings)}, nil
}

// RemoveMesh removes a mesh without nodes, with its settings, records,
// versions, policies and audit entries; the removal is the control
// plane's audit entry.
func (s *Nodes) RemoveMesh(ctx context.Context, by Actor, name string) error {
	if _, err := validate.Name("mesh", name); err != nil {
		return invalid(err)
	}
	return s.db.Tx(ctx, func(raw *store.Tx) error {
		m, err := PickMesh(raw, name)
		if err != nil {
			return err
		}
		nodes, err := raw.In(m.ID).Nodes()
		if err != nil {
			return err
		}
		if len(nodes) > 0 {
			names := make([]string, len(nodes))
			for i, n := range nodes {
				names[i] = n.Name
			}
			return errorf(http.StatusConflict, proto.CodeConflict, "mesh %q still has %d node(s) (%s): remove them first (`meshctl node rm --mesh %s <name>`)",
				name, len(nodes), strings.Join(names, ", "), name)
		}
		if err := raw.DeleteMesh(m.ID); err != nil {
			return err
		}
		return raw.Audit(by.Name, by.UserID, "mesh.rm", name, map[string]any{"uuid": m.UUID, "slot": m.Slot})
	})
}

// settings checks req and returns the mutation that writes it.
func (s *Nodes) settings(req SettingsRequest) (func(*store.MeshTx, *Change) error, error) {
	values := map[string]string{}
	if req.DefaultMTU != nil {
		if _, err := validate.IntRange("default_mtu", *req.DefaultMTU, proto.MinMTU, proto.MaxMTU); err != nil {
			return nil, invalid(err)
		}
		values[store.SettingDefaultMTU] = strconv.Itoa(*req.DefaultMTU)
	}
	if req.Upstream != nil {
		for _, a := range *req.Upstream {
			if _, err := validate.IP("upstream_dns", a); err != nil {
				return nil, invalid(err)
			}
		}
		if n := len(*req.Upstream); n == 0 || n > 4 {
			return nil, invalid(&validate.Error{Field: "upstream_dns", Reason: "give one to four IPv4 or IPv6 addresses, comma separated"})
		}
		values[store.SettingUpstreamDNS] = strings.Join(*req.Upstream, ",")
	}
	// How the zone and the extra domains sit with each other the reconciler
	// checks, which knows the zone in effect.
	if req.Zone != nil {
		if _, err := validate.Zone("dns_zone", *req.Zone); err != nil {
			return nil, invalid(err)
		}
		values[store.SettingDNSZone] = *req.Zone
	}
	if req.Domains != nil {
		ds, err := validate.Domains("dns_extra_domains", "", *req.Domains)
		if err == nil {
			err = CheckServerDomains(s.serverURL, ds)
		}
		if err != nil {
			return nil, invalid(err)
		}
		values[store.SettingDNSExtraDomains] = strings.Join(ds, ",")
	}
	var cidr netip.Prefix
	if req.MeshCIDR != nil {
		var err error
		if cidr, err = meshCIDR(*req.MeshCIDR); err != nil {
			return nil, invalid(err)
		}
	}
	if req.DefaultHub != nil && *req.DefaultHub != "" {
		if _, err := validate.Name("default_hub", *req.DefaultHub); err != nil {
			return nil, invalid(err)
		}
	}
	return func(tx *store.MeshTx, c *Change) error {
		if req.DefaultHub != nil {
			values[store.SettingDefaultHub] = ""
			if hub := *req.DefaultHub; hub != "" {
				n, err := tx.NodeByName(hub)
				if err != nil || n.Role != string(proto.RoleServer) {
					return invalid(&validate.Error{Field: "default_hub", Value: hub, Reason: "must be a server of this mesh"})
				}
				values[store.SettingDefaultHub] = strconv.FormatInt(n.ID, 10)
			}
		}
		if cidr.IsValid() {
			cur, err := tx.Setting(store.SettingMeshCIDR)
			if err != nil {
				return err
			}
			if cidr.String() != cur {
				nodes, err := tx.Nodes()
				if err != nil {
					return err
				}
				if len(nodes) > 0 {
					return invalid(&validate.Error{Field: "mesh_cidr", Reason: "can only change while the mesh has no nodes"})
				}
				values[store.SettingMeshCIDR] = cidr.String()
			}
		}
		for _, k := range slices.Sorted(maps.Keys(values)) {
			if err := tx.SetSetting(k, values[k]); err != nil {
				return err
			}
		}
		c.Detail = values
		return nil
	}, nil
}

// meshCIDR checks a mesh's IPv4 network.
func meshCIDR(s string) (netip.Prefix, error) {
	p, err := validate.Prefix4("mesh_cidr", s)
	if err != nil {
		return netip.Prefix{}, err
	}
	if err := planner.CheckMeshCIDR(p); err != nil {
		return netip.Prefix{}, &validate.Error{Field: "mesh_cidr", Value: s, Reason: err.Error()}
	}
	return p, nil
}

// meshInfo is m as the admin API shows it; tx is scoped to m.
func meshInfo(tx *store.MeshTx, m store.Mesh) (MeshInfo, error) {
	settings, err := tx.Settings()
	if err != nil {
		return MeshInfo{}, err
	}
	nodes, err := tx.Nodes()
	if err != nil {
		return MeshInfo{}, err
	}
	out := MeshInfo{Name: m.Name, UUID: m.UUID, Slot: m.Slot, CreatedAt: m.CreatedAt, Nodes: len(nodes),
		MeshCIDR: settings[store.SettingMeshCIDR], Zone: planner.Input{Zone: settings[store.SettingDNSZone]}.ZoneName(),
		Domains: store.SplitDomains(settings[store.SettingDNSExtraDomains])}
	if v := settings[store.SettingMeshCIDR6]; v != store.MeshCIDR6Off {
		out.MeshCIDR6 = v
	}
	for _, f := range strings.Split(settings[store.SettingUpstreamDNS], ",") {
		if f = strings.TrimSpace(f); f != "" {
			out.Upstream = append(out.Upstream, f)
		}
	}
	out.DefaultMTU, _ = strconv.Atoi(settings[store.SettingDefaultMTU]) // validated on the way in
	for _, n := range nodes {
		if strconv.FormatInt(n.ID, 10) == settings[store.SettingDefaultHub] {
			out.DefaultHub = n.Name
		}
	}
	return out, nil
}

// freeSlot is the lowest slot no mesh has.
func freeSlot(meshes []store.Mesh) (int, error) {
	for slot := 0; slot <= store.MaxSlot; slot++ {
		if !slices.ContainsFunc(meshes, func(m store.Mesh) bool { return m.Slot == slot }) {
			return slot, nil
		}
	}
	return 0, errorf(http.StatusConflict, proto.CodeConflict,
		"every slot is taken: a control plane holds at most %d meshes; remove one first", store.MaxSlot+1)
}

// freeCIDR is the first 10.x.0.0/16 that no mesh overlaps, from
// 10.77.0.0/16 (the first mesh's) up, then below it.
func freeCIDR(spaces []meshSpace) (netip.Prefix, error) {
	for i := range 255 {
		x := 77 + i
		if x > 255 {
			x -= 255
		}
		p := netip.PrefixFrom(netip.AddrFrom4([4]byte{10, byte(x), 0, 0}), 16)
		if !slices.ContainsFunc(spaces, func(sp meshSpace) bool { return sp.v4.IsValid() && sp.v4.Overlaps(p) }) {
			return p, nil
		}
	}
	return netip.Prefix{}, errorf(http.StatusConflict, proto.CodeConflict, "no 10.x.0.0/16 is free: give the mesh's network (--mesh-cidr)")
}

// meshSpace is what of a mesh no other mesh of the control plane may share:
// its networks and its DNS domains (the zone first, then the extra ones).
type meshSpace struct {
	name    string
	v4, v6  netip.Prefix
	domains []string
}

// spaceOf reads tx's mesh's space.
func spaceOf(tx *store.MeshTx, name string) (meshSpace, error) {
	settings, err := tx.Settings()
	if err != nil {
		return meshSpace{}, err
	}
	sp := meshSpace{name: name, domains: []string{planner.Input{Zone: settings[store.SettingDNSZone]}.ZoneName()}}
	sp.domains = append(sp.domains, store.SplitDomains(settings[store.SettingDNSExtraDomains])...)
	sp.v4, _ = netip.ParsePrefix(settings[store.SettingMeshCIDR]) // validated on the way in
	if v := settings[store.SettingMeshCIDR6]; v != store.MeshCIDR6Off {
		sp.v6, _ = netip.ParsePrefix(v)
	}
	return sp, nil
}

func spacesOf(tx *store.Tx, meshes []store.Mesh) ([]meshSpace, error) {
	out := make([]meshSpace, 0, len(meshes))
	for _, m := range meshes {
		sp, err := spaceOf(tx.In(m.ID), m.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, sp)
	}
	return out, nil
}

// clash is what a shares with b, "" for nothing.
func clash(a, b meshSpace) string {
	switch {
	case a.v4.IsValid() && b.v4.IsValid() && a.v4.Overlaps(b.v4):
		return fmt.Sprintf("its IPv4 network %s overlaps mesh %s's %s", a.v4, b.name, b.v4)
	case a.v6.IsValid() && b.v6.IsValid() && a.v6.Overlaps(b.v6):
		return fmt.Sprintf("its IPv6 prefix %s overlaps mesh %s's %s", a.v6, b.name, b.v6)
	}
	for _, d := range a.domains {
		for _, e := range b.domains {
			if d == e || strings.HasSuffix(d, "."+e) || strings.HasSuffix(e, "."+d) {
				return fmt.Sprintf("its DNS domain %s is mesh %s's %s, inside it or around it", d, b.name, e)
			}
		}
	}
	return ""
}

// checkApart refuses tx's mesh when it shares a network or a DNS domain
// with another mesh: agents of two meshes may share a host, its routes and
// its resolver.
func checkApart(tx *store.MeshTx) error {
	meshes, err := tx.Meshes()
	if err != nil {
		return err
	}
	spaces, err := spacesOf(tx.Tx, meshes)
	if err != nil {
		return err
	}
	i := slices.IndexFunc(meshes, func(m store.Mesh) bool { return m.ID == tx.MeshID() })
	if i < 0 {
		return fmt.Errorf("mesh %d: %w", tx.MeshID(), store.ErrNotFound)
	}
	for j, other := range spaces {
		if j == i {
			continue
		}
		if why := clash(spaces[i], other); why != "" {
			return errorf(http.StatusConflict, proto.CodeConflict, "mesh %s: %s; meshes of one control plane share no network and no DNS domain",
				spaces[i].name, why)
		}
	}
	return nil
}

func (a *api) adminMeshes(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.Meshes(r.Context())
	a.reply(w, r, out, err)
}

func (a *api) adminAddMesh(w http.ResponseWriter, r *http.Request) {
	var req MeshAddRequest
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.AddMesh(r.Context(), cli(), req)
	a.reply(w, r, out, err)
}

func (a *api) adminShowMesh(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.ShowMesh(r.Context(), r.PathValue("name"))
	a.reply(w, r, out, err)
}

func (a *api) adminSetMesh(w http.ResponseWriter, r *http.Request) {
	var req MeshSetRequest
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.SetMesh(r.Context(), cli(), r.PathValue("name"), req)
	a.reply(w, r, out, err)
}

func (a *api) adminRemoveMesh(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	err := a.nodes.RemoveMesh(r.Context(), cli(), name)
	a.reply(w, r, map[string]string{"removed": name}, err)
}

func meshPath(name string) string { return AdminMeshes + "/" + url.PathEscape(name) }

// Meshes lists the control plane's meshes.
func (c *AdminClient) Meshes(ctx context.Context) (MeshList, error) {
	var out MeshList
	err := c.do(ctx, http.MethodGet, AdminMeshes, nil, &out)
	return out, err
}

// AddMesh creates a mesh.
func (c *AdminClient) AddMesh(ctx context.Context, req MeshAddRequest) (MeshUpdate, error) {
	var out MeshUpdate
	err := c.do(ctx, http.MethodPost, AdminMeshes, req, &out)
	return out, err
}

// ShowMesh returns one mesh.
func (c *AdminClient) ShowMesh(ctx context.Context, name string) (MeshInfo, error) {
	var out MeshInfo
	err := c.do(ctx, http.MethodGet, meshPath(name), nil, &out)
	return out, err
}

// SetMesh renames a mesh or changes its settings.
func (c *AdminClient) SetMesh(ctx context.Context, name string, req MeshSetRequest) (MeshUpdate, error) {
	var out MeshUpdate
	err := c.do(ctx, http.MethodPatch, meshPath(name), req, &out)
	return out, err
}

// RemoveMesh removes a mesh without nodes.
func (c *AdminClient) RemoveMesh(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodDelete, meshPath(name), nil, nil)
}
