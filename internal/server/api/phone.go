package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/netip"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/phone"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Phones have no agent (spec S3). The server makes a phone's key pair when
// it shows the phone's profile and keeps only the public key (amendment A6),
// so each profile replaces the one before it. What the server knows of a
// phone afterwards is what its servers report: the handshakes they see.

func (a *api) adminPhone(w http.ResponseWriter, r *http.Request) {
	var req PhoneRequest
	if err := a.readJSON(w, r, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.Phone(r.Context(), cli(), r.PathValue("name"), req)
	a.reply(w, r, out, err)
}

// reachable is a server the planner puts in other nodes' peer lists: joined,
// enabled, with an endpoint (amendment A17).
func reachable(n store.Node) bool {
	return n.Role == string(proto.RoleServer) && n.Enrolled && !n.Disabled && n.WGPubKey != "" && n.Endpoint != ""
}

// Phone makes a phone's WireGuard profile with a new key pair. An empty
// req.Exit is the mesh-only profile; a server that can be an exit takes the
// phone's internet (the via-exit profile). The answer is the only place the
// private key ever is: it is not stored, logged or audited.
func (s *Nodes) Phone(ctx context.Context, by Actor, name string, req PhoneRequest) (PhoneProfile, error) {
	name, err := checkName(name)
	if err != nil {
		return PhoneProfile{}, err
	}
	if req.Exit != "" {
		if _, err := validate.Name("exit", req.Exit); err != nil {
			return PhoneProfile{}, invalid(err)
		}
	}
	priv, pub, err := phone.NewKey(rand.Reader)
	if err != nil {
		return PhoneProfile{}, err
	}
	var id int64
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "phone.profile", Target: name}, func(tx *store.Tx, c *Change) error {
		n, err := nodeByName(tx, name, "node")
		if err != nil {
			return err
		}
		switch {
		case n.Role != string(proto.RolePhone):
			return errorf(http.StatusConflict, proto.CodeConflict, "%s is a %s; only phones get a profile (other nodes join with `meshctl agent join`)", name, n.Role)
		case n.Disabled:
			return errorf(http.StatusConflict, proto.CodeConflict, "phone %q is disabled; enable it first", name)
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(nodes, reachable) {
			return errorf(http.StatusConflict, proto.CodeConflict,
				"no server has joined with an endpoint yet; a phone's profile lists the servers it dials, so join one first")
		}
		if req.Exit != "" {
			ex, err := nodeByName(tx, req.Exit, "exit")
			if err != nil {
				return err
			}
			if !reachable(ex) || !ex.CanExit {
				return errorf(http.StatusConflict, proto.CodeConflict,
					"%s cannot be this phone's exit: it must be a joined, enabled server with an endpoint that can be an exit (`meshctl node set %s --can-exit`)",
					ex.Name, ex.Name)
			}
		}
		id = n.ID
		c.Detail = map[string]any{"exit": req.Exit, "rekey": n.Enrolled}
		n.WGPubKey, n.Enrolled = pub, true
		return tx.UpdateNode(n)
	})
	if err != nil {
		return PhoneProfile{}, err
	}

	out := PhoneProfile{Version: res.Version, Exit: req.Exit, Warnings: problemStrings(res.Warnings)}
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		self, cfg, err := configFor(tx, id)
		if err != nil {
			return err
		}
		if cfg == nil {
			return fmt.Errorf("no config was computed for phone %s; the node list shows the plan's warnings", name)
		}
		hubName, ok := phone.Hub(*cfg)
		if !ok {
			return fmt.Errorf("the config of phone %s has no peer that carries the mesh", name)
		}
		hub, err := nodeByName(tx, hubName, "hub")
		if err != nil {
			return err
		}
		// The hub's resolver over the phone's families (M8): IPv4 first.
		dns, dns6, err := phoneDNS(*cfg, hub)
		if err != nil {
			return err
		}
		if out.Profile, err = phone.Render(phone.Profile{Config: *cfg, PrivateKey: priv, DNS: dns, DNS6: dns6, Exit: req.Exit}); err != nil {
			return err
		}
		// The profile the phone holds: when its config changes later, the
		// node shows the profile is out of date.
		hash, err := cfg.Hash()
		if err != nil {
			return err
		}
		if err := tx.SetApplied(id, hash, self.ConfigVersion); err != nil {
			return err
		}
		if self, err = tx.NodeByID(id); err != nil {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		out.Node = nodeInfo(self, nodes)
		return withHandshake(tx, &out.Node)
	})
	if err != nil {
		return PhoneProfile{}, fmt.Errorf("make the profile of phone %s (its new key is registered; make another profile): %w", name, err)
	}
	return out, nil
}

// withHandshake fills in, for a phone, the latest handshake its servers
// reported.
func withHandshake(tx *store.Tx, info *NodeInfo) error {
	if info.Role != string(proto.RolePhone) {
		return nil
	}
	var err error
	info.LastHandshake, err = tx.LastHandshakeWith(info.ID)
	return err
}

// phoneDNS is the hub's resolver as a phone's profile names it: the hub's
// IPv4 mesh IP for a phone with IPv4, else its IPv6 one, and the IPv6 one
// beside it for a phone with both (M8).
func phoneDNS(cfg proto.NodeConfig, hub store.Node) (netip.Addr, netip.Addr, error) {
	v4, err4 := netip.ParseAddr(hub.MeshIP)
	v6, err6 := netip.ParseAddr(hub.MeshIP6)
	has4, has6 := cfg.Interface.Address != "" && err4 == nil, cfg.Interface.Address6 != "" && err6 == nil
	switch {
	case has4 && has6:
		return v4, v6, nil
	case has4:
		return v4, netip.Addr{}, nil
	case has6:
		return v6, netip.Addr{}, nil
	}
	return netip.Addr{}, netip.Addr{}, fmt.Errorf("hub %s has no mesh IP in a family the phone %s has", hub.Name, cfg.Node.Name)
}
