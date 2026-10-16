package proto

import (
	"errors"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The mesh view (spec S7.6): every node as the server's dashboard shows it,
// for the local UI on each node. It leaves out what one node has no
// business knowing about another: keys, endpoints, hostnames and error
// text.

// A node's dashboard colour: green, amber, red or grey.
const (
	HealthOK   = "ok"
	HealthWarn = "warn"
	HealthBad  = "bad"
	HealthOff  = "off"
)

// Healths are the values of MeshNode.Health.
var Healths = []string{HealthOK, HealthWarn, HealthBad, HealthOff}

// MeshView is the answer to GET mesh.
type MeshView struct {
	// Version is the mesh's latest config version.
	Version    int64      `json:"version"`
	ServerTime int64      `json:"server_time"`
	Nodes      []MeshNode `json:"nodes"`
}

// MeshNode is one node of the mesh view.
type MeshNode struct {
	Name    string `json:"name"`
	Role    Role   `json:"role"`
	MeshIP  string `json:"mesh_ip"`
	Hub     string `json:"hub,omitempty"`
	Exit    string `json:"exit,omitempty"`
	CanExit bool   `json:"can_exit,omitempty"`
	// Health and State are the dashboard's colour and label ("ok", "in
	// sync").
	Health string `json:"health"`
	State  string `json:"state"`
	// LastSeen is the node's last heartbeat; for a phone, the last
	// handshake a server saw.
	LastSeen       int64 `json:"last_seen,omitempty"`
	ConfigVersion  int64 `json:"config_version"`
	AppliedVersion int64 `json:"applied_version"`
}

// maxMeshNodes caps a mesh view: the mesh is at most a /16.
const maxMeshNodes = 1 << 16

// Validate checks a mesh view from the server before the agent caches or
// shows it, and reports every bad field.
func (v *MeshView) Validate() error {
	if len(v.Nodes) > maxMeshNodes {
		return fmt.Errorf("nodes: at most %d entries (got %d)", maxMeshNodes, len(v.Nodes))
	}
	var errs []error
	if v.Version < 0 || v.ServerTime < 0 {
		errs = append(errs, fmt.Errorf("version and server_time must not be negative (got %d, %d)", v.Version, v.ServerTime))
	}
	for i, n := range v.Nodes {
		field := fmt.Sprintf("nodes[%d]", i)
		if _, err := validate.Name(field+".name", n.Name); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.Enum(field+".role", string(n.Role), string(RoleServer), string(RoleSpoke), string(RolePhone)); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.IPv4(field+".mesh_ip", n.MeshIP); err != nil {
			errs = append(errs, err)
		}
		if n.Hub != "" {
			if _, err := validate.Name(field+".hub", n.Hub); err != nil {
				errs = append(errs, err)
			}
		}
		if n.Exit != "" {
			if _, err := validate.Name(field+".exit", n.Exit); err != nil {
				errs = append(errs, err)
			}
		}
		if _, err := validate.Enum(field+".health", n.Health, Healths...); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.Text(field+".state", n.State, 200); err != nil {
			errs = append(errs, err)
		}
		if n.LastSeen < 0 || n.ConfigVersion < 0 || n.AppliedVersion < 0 {
			errs = append(errs, errors.New(field+": times and versions must not be negative"))
		}
	}
	return errors.Join(errs...)
}
