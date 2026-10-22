package api

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The mesh's IPv6 prefix (M8). Setting it gives every node that may have
// IPv6 (ip_family dual or ipv6) an IPv6 mesh address in one change;
// giving it up takes every IPv6 mesh address and container /112 away. A
// prefix that holds addresses is not swapped for another: give IPv6 up
// first. The admin API, `meshctl ipv6` and the web UI's Settings page use
// it.

// AdminIPv6 is the admin API's IPv6 prefix.
const AdminIPv6 = "/admin/v1/ipv6"

// The words IPv6Request.Prefix takes besides a prefix.
const (
	IPv6Auto = "auto"
	IPv6Off  = "off"
)

// IPv6Request is PUT /admin/v1/ipv6: Prefix is an IPv6 /48 to /64, IPv6Auto
// (a random unique local /64, or the prefix there is) or IPv6Off (or empty)
// for none.
type IPv6Request struct {
	Prefix string `json:"prefix"`
}

// IPv6View is GET /admin/v1/ipv6 and the answer to a change: the prefix
// ("" for none) and how many nodes have an IPv6 mesh address.
type IPv6View struct {
	Prefix    string `json:"prefix,omitempty"`
	Addressed int    `json:"addressed"`
	// Version is the config version a change made; 0 when none.
	Version  int64    `json:"version,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ULAPrefix is a random unique local IPv6 /64 (RFC 4193): fd, a 40-bit
// global ID from random, subnet 0.
func ULAPrefix(random func([]byte) (int, error)) (netip.Prefix, error) {
	var a [16]byte
	a[0] = 0xfd
	if _, err := random(a[1:6]); err != nil {
		return netip.Prefix{}, fmt.Errorf("draw a random IPv6 prefix: %w", err)
	}
	return netip.PrefixFrom(netip.AddrFrom16(a), 64), nil
}

// IPv6 returns the mesh's IPv6 prefix.
func (s *Nodes) IPv6(ctx context.Context) (IPv6View, error) {
	var out IPv6View
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		v, err := ipv6View(tx)
		out = v
		return err
	})
	return out, err
}

func ipv6View(tx *store.Tx) (IPv6View, error) {
	var v IPv6View
	cur, err := tx.Setting(store.SettingMeshCIDR6)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return v, err
	}
	nodes, err := tx.Nodes()
	if err != nil {
		return v, err
	}
	v.Prefix = cur
	for _, n := range nodes {
		if n.MeshIP6 != "" {
			v.Addressed++
		}
	}
	return v, nil
}

// SetIPv6 sets, generates or gives up the mesh's IPv6 prefix.
func (s *Nodes) SetIPv6(ctx context.Context, by Actor, req IPv6Request) (IPv6View, error) {
	want := strings.TrimSpace(req.Prefix)
	var p netip.Prefix
	switch want {
	case "", IPv6Off:
		want = IPv6Off
	case IPv6Auto:
	default:
		var err error
		if p, err = validate.Prefix6("prefix", want); err == nil {
			err = planner.CheckMeshCIDR6(p)
		}
		if err != nil {
			return IPv6View{}, invalid(err)
		}
	}
	var out IPv6View
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "ipv6.set", Target: want}, func(tx *store.Tx, c *Change) error {
		cur, err := tx.Setting(store.SettingMeshCIDR6)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		if want == IPv6Auto {
			if p, err = netip.ParsePrefix(cur); err != nil {
				if p, err = ULAPrefix(rand.Read); err != nil {
					return err
				}
			}
		}
		if want == IPv6Off {
			return giveUpIPv6(tx, nodes, c)
		}
		if cur != "" && cur != p.String() {
			for _, n := range nodes {
				if n.MeshIP6 != "" {
					return errorf(http.StatusConflict, proto.CodeConflict,
						"the mesh's IPv6 prefix is %s and %s has an address in it; give IPv6 up first (`meshctl ipv6 off`), then set the new prefix", cur, n.Name)
				}
			}
		}
		if err := tx.SetSetting(store.SettingMeshCIDR6, p.String()); err != nil {
			return err
		}
		in, _, err := LoadInput(tx)
		if err != nil {
			return err
		}
		used, added := usedAddrs(nodes, 0), 0
		for _, n := range nodes {
			if n.Family == proto.FamilyIPv4 || n.MeshIP6 != "" {
				continue
			}
			if err := addIPv6(in, &n, netip.Addr{}, blockOf(in, n), used); err != nil {
				return err
			}
			if err := tx.UpdateNode(n); err != nil {
				return err
			}
			added++
		}
		c.Target, c.Detail = p.String(), map[string]any{"prefix": p.String(), "addressed": added}
		return nil
	})
	if err != nil {
		return IPv6View{}, err
	}
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		out, err = ipv6View(tx)
		return err
	})
	out.Version, out.Warnings = res.Version, problemStrings(res.Warnings)
	return out, err
}

// giveUpIPv6 takes the mesh's IPv6 away: the prefix, every IPv6 mesh
// address and container /112. An IPv6-only node would have no address
// left, so it must change first.
func giveUpIPv6(tx *store.Tx, nodes []store.Node, c *Change) error {
	for _, n := range nodes {
		if n.Family == proto.FamilyIPv6 {
			return errorf(http.StatusConflict, proto.CodeConflict,
				"%s is IPv6 only; make it dual or ipv4 first (`meshctl node set %s --family dual`)", n.Name, n.Name)
		}
	}
	removed := 0
	for _, n := range nodes {
		if n.MeshIP6 == "" && n.ContainerSubnet6 == "" {
			continue
		}
		n.MeshIP6, n.ContainerSubnet6 = "", ""
		if err := tx.UpdateNode(n); err != nil {
			return err
		}
		removed++
	}
	if err := tx.DeleteSetting(store.SettingMeshCIDR6); err != nil {
		return err
	}
	c.Detail = map[string]any{"prefix": "", "removed": removed}
	return nil
}

func (a *api) adminIPv6(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.IPv6(r.Context())
	a.reply(w, r, out, err)
}

func (a *api) adminSetIPv6(w http.ResponseWriter, r *http.Request) {
	var req IPv6Request
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.SetIPv6(r.Context(), cli(), req)
	a.reply(w, r, out, err)
}

// IPv6 returns the mesh's IPv6 prefix.
func (c *AdminClient) IPv6(ctx context.Context) (IPv6View, error) {
	var out IPv6View
	err := c.do(ctx, http.MethodGet, AdminIPv6, nil, &out)
	return out, err
}

// SetIPv6 sets, generates or gives up the mesh's IPv6 prefix.
func (c *AdminClient) SetIPv6(ctx context.Context, req IPv6Request) (IPv6View, error) {
	var out IPv6View
	err := c.do(ctx, http.MethodPut, AdminIPv6, req, &out)
	return out, err
}
