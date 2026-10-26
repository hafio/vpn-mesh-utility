package proto

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Agent API paths (spec S6). Everything under AgentPrefix must stay reachable
// from every node's country if a GeoIP allowlist fronts the UI.
const (
	AgentPrefix = "/api/v1/agent/"
	PathJoin    = "/api/v1/agent/join"
	PathConfig  = "/api/v1/agent/config"
	PathEvents  = "/api/v1/agent/events"
	PathStatus  = "/api/v1/agent/status"
	PathPrefs   = "/api/v1/agent/prefs"
	PathMesh    = "/api/v1/agent/mesh"
)

// Public paths served next to the agent API: the health probe and the
// self-hosted install script and binaries (amendment A14).
const (
	PathHealth   = "/healthz"
	PathInstall  = "/install.sh"
	PathDownload = "/dl/"
)

// ETag is the HTTP entity tag of a config: its quoted Hash. GET config with
// If-None-Match set to the applied config's ETag answers 304 when nothing
// changed.
func ETag(hash string) string { return `"` + hash + `"` }

// MaxBodyBytes caps every agent request body.
const MaxBodyBytes = 256 << 10

// Error codes carried in ErrorBody.Code. Agents branch on these, never on the
// message text.
const (
	CodeBadRequest   = "bad_request"
	CodeUnauthorized = "unauthorized"
	CodeClockSkew    = "clock_skew"
	CodeReplay       = "replay"
	CodeNodeDisabled = "node_disabled"
	CodeTokenInvalid = "token_invalid"
	CodeNotFound     = "not_found"
	CodeConflict     = "conflict"
	CodeRateLimited  = "rate_limited"
	CodeTooLarge     = "too_large"
	CodeInternal     = "internal"
)

// ErrorBody is the JSON body of every non-2xx agent API response. Message is
// actionable and never echoes a secret; ServerTime accompanies clock_skew so
// the operator can see how far off the node is.
type ErrorBody struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	ServerTime int64  `json:"server_time,omitempty"`
}

// Values an agent may report about its platform.
var (
	OSes      = []string{"linux", "windows"}
	Arches    = []string{"amd64", "arm64"}
	Backends  = []string{"kernel", "userspace"}
	reHexHash = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// JoinRequest enrolls a node with a one-time token (spec S6.1). It is the one
// agent request authenticated by the token rather than a signature: the node
// has no registered identity key yet.
type JoinRequest struct {
	Token             string `json:"token"`
	Name              string `json:"name,omitempty"`
	Hostname          string `json:"hostname"`
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	WGBackend         string `json:"wg_backend"`
	WGPublicKey       string `json:"wg_public_key"`
	IdentityPublicKey string `json:"identity_public_key"`
	// Endpoint is the IPv4:port given with --endpoint. PublicIP is the public
	// IPv4 the agent found on its own interfaces. A server node's endpoint is
	// Endpoint, else PublicIP with the node's listen port, else the address
	// the request came from (amendment A17); spokes ignore both.
	Endpoint     string `json:"endpoint,omitempty"`
	PublicIP     string `json:"public_ip,omitempty"`
	AgentVersion string `json:"agent_version"`
	// The IPv6 side (M8): Endpoint6 is the [IPv6]:port given with
	// --endpoint6, PublicIP6 the public IPv6 the agent found; a server's
	// IPv6 endpoint is chosen from them as its IPv4 one is.
	Endpoint6 string `json:"endpoint6,omitempty"`
	PublicIP6 string `json:"public_ip6,omitempty"`
	// MachineID is the host's machine id as HelloRequest carries it
	// (amendment A50): the node is pinned to this machine.
	MachineID string `json:"machine_id"`
}

// Validate normalises the request in place and reports every bad field.
func (r *JoinRequest) Validate() error {
	var errs []error
	keep := func(s string, err error, dst *string) {
		if err != nil {
			errs = append(errs, err)
			return
		}
		*dst = s
	}
	s, err := validate.Token("token", r.Token)
	keep(s, err, &r.Token)
	if r.Name != "" {
		s, err = validate.Name("name", r.Name)
		keep(s, err, &r.Name)
	}
	s, err = validate.Hostname("hostname", r.Hostname)
	keep(s, err, &r.Hostname)
	s, err = validate.Enum("os", r.OS, OSes...)
	keep(s, err, &r.OS)
	s, err = validate.Enum("arch", r.Arch, Arches...)
	keep(s, err, &r.Arch)
	s, err = validate.Enum("wg_backend", r.WGBackend, Backends...)
	keep(s, err, &r.WGBackend)
	s, err = validate.Key32("wg_public_key", r.WGPublicKey)
	keep(s, err, &r.WGPublicKey)
	s, err = validate.Key32("identity_public_key", r.IdentityPublicKey)
	keep(s, err, &r.IdentityPublicKey)
	if r.Endpoint != "" {
		if ap, err := validate.Endpoint("endpoint", r.Endpoint); err != nil {
			errs = append(errs, err)
		} else {
			r.Endpoint = ap.String()
		}
	}
	if r.PublicIP != "" {
		if ip, err := validate.IPv4("public_ip", r.PublicIP); err != nil {
			errs = append(errs, err)
		} else {
			r.PublicIP = ip.String()
		}
	}
	if r.Endpoint6 != "" {
		if ap, err := validate.Endpoint6("endpoint6", r.Endpoint6); err != nil {
			errs = append(errs, err)
		} else {
			r.Endpoint6 = ap.String()
		}
	}
	if r.PublicIP6 != "" {
		if ip, err := validate.IPv6("public_ip6", r.PublicIP6); err != nil {
			errs = append(errs, err)
		} else {
			r.PublicIP6 = ip.String()
		}
	}
	s, err = validate.Version("agent_version", r.AgentVersion)
	keep(s, err, &r.AgentVersion)
	if !ValidSecret(r.MachineID) {
		errs = append(errs, errors.New("machine_id must be 64 lowercase hex characters"))
	}
	return errors.Join(errs...)
}

// JoinResponse tells the agent who it is now.
type JoinResponse struct {
	NodeID        int64  `json:"node_id"`
	Name          string `json:"name"`
	Role          Role   `json:"role"`
	MeshIP        string `json:"mesh_ip"`
	ConfigVersion int64  `json:"config_version"`
	ServerTime    int64  `json:"server_time"`
	// MeshIP6 is the node's IPv6 mesh address, when it has one (M8).
	MeshIP6 string `json:"mesh_ip6,omitempty"`
	// Mesh, MeshUUID and Slot are the mesh the node joined (amendment
	// A49); the slot picks the names its agent uses on the host.
	Mesh     string `json:"mesh,omitempty"`
	MeshUUID string `json:"mesh_uuid,omitempty"`
	Slot     int    `json:"slot,omitempty"`
	// ConfigKey is the control plane's public key (base64 Ed25519), which
	// the agent pins: every config, mesh view and definitive refusal is
	// signed with it (amendment A50).
	ConfigKey string `json:"config_key"`
}

// StatusReport is the agent heartbeat (spec S6.3): every 30 s and right after
// an apply.
type StatusReport struct {
	AppliedVersion int64        `json:"applied_version"`
	AppliedHash    string       `json:"applied_hash,omitempty"`
	AgentVersion   string       `json:"agent_version"`
	WGBackend      string       `json:"wg_backend"`
	Peers          []PeerStatus `json:"peers"`
	// Errors are apply or runtime failures, newest first.
	Errors []string `json:"errors,omitempty"`
	// Degraded names optional apply steps that failed while the tunnel stayed
	// up ("firewall", "dns").
	Degraded []string  `json:"degraded,omitempty"`
	Rollback *Rollback `json:"rollback,omitempty"`
	// DNS is the node's split DNS: one of DNSStates, empty when its config
	// has no DNS section.
	DNS string `json:"dns,omitempty"`
	// Families are the address families the node's host can use (M8):
	// FamilyIPv4, FamilyIPv6 or both. Empty is "not said", which the server
	// reads as every family the node has; a family left out is planned
	// away from the node until it is reported again.
	Families []string `json:"families,omitempty"`
}

// PeerStatus is the runtime state of one peer as the agent sees it.
type PeerStatus struct {
	Name          string `json:"name"`
	LastHandshake int64  `json:"last_handshake,omitempty"`
	RxBytes       int64  `json:"rx_bytes"`
	TxBytes       int64  `json:"tx_bytes"`
	Endpoint      string `json:"endpoint,omitempty"`
}

// Rollback reports that safe-apply reverted a config (spec S6.4).
type Rollback struct {
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Reason string `json:"reason"`
}

// Limits on a status report, so one node cannot bloat the database.
const (
	maxStatusPeers  = 1024
	maxStatusErrors = 32
)

// Validate checks a status report and reports every bad field.
func (r *StatusReport) Validate() error {
	var errs []error
	if r.AppliedVersion < 0 {
		errs = append(errs, fmt.Errorf("applied_version: must not be negative (got %d)", r.AppliedVersion))
	}
	if r.AppliedHash != "" && !reHexHash.MatchString(r.AppliedHash) {
		errs = append(errs, errors.New("applied_hash: must be 64 lowercase hex chars"))
	}
	if _, err := validate.Version("agent_version", r.AgentVersion); err != nil {
		errs = append(errs, err)
	}
	if _, err := validate.Enum("wg_backend", r.WGBackend, Backends...); err != nil {
		errs = append(errs, err)
	}
	if len(r.Peers) > maxStatusPeers {
		errs = append(errs, fmt.Errorf("peers: at most %d entries (got %d)", maxStatusPeers, len(r.Peers)))
	} else {
		for i, p := range r.Peers {
			field := fmt.Sprintf("peers[%d]", i)
			if _, err := validate.Name(field+".name", p.Name); err != nil {
				errs = append(errs, err)
			}
			if p.LastHandshake < 0 || p.RxBytes < 0 || p.TxBytes < 0 {
				errs = append(errs, fmt.Errorf("%s: counters must not be negative", field))
			}
			if p.Endpoint != "" {
				if _, err := netip.ParseAddrPort(p.Endpoint); err != nil {
					errs = append(errs, fmt.Errorf("%s.endpoint: must be ip:port (got %q)", field, p.Endpoint))
				}
			}
		}
	}
	if len(r.Errors) > maxStatusErrors {
		errs = append(errs, fmt.Errorf("errors: at most %d entries (got %d)", maxStatusErrors, len(r.Errors)))
	} else {
		for i, e := range r.Errors {
			if _, err := validate.Text(fmt.Sprintf("errors[%d]", i), e, 1024); err != nil {
				errs = append(errs, err)
			}
		}
	}
	for i, d := range r.Degraded {
		if _, err := validate.Enum(fmt.Sprintf("degraded[%d]", i), d, DegradableSteps...); err != nil {
			errs = append(errs, err)
		}
	}
	if r.Rollback != nil {
		if _, err := validate.Text("rollback.reason", r.Rollback.Reason, 1024); err != nil {
			errs = append(errs, err)
		}
	}
	if r.DNS != "" {
		if _, err := validate.Enum("dns", r.DNS, DNSStates...); err != nil {
			errs = append(errs, err)
		}
	}
	seen := map[string]bool{}
	for i, f := range r.Families {
		field := fmt.Sprintf("families[%d]", i)
		if _, err := validate.Enum(field, f, FamilyIPv4, FamilyIPv6); err != nil {
			errs = append(errs, err)
		} else if seen[f] {
			errs = append(errs, fmt.Errorf("%s: %s is listed twice", field, f))
		}
		seen[f] = true
	}
	return errors.Join(errs...)
}

// PrefsRequest changes a node's own settings (amendment A10): the exit and
// keepalive of a spoke, the MTU of any node. A nil field is left alone; an
// Exit of "" turns exit mode off. The server applies it like an operator's
// `meshctl node set`, so it is versioned and audited.
type PrefsRequest struct {
	Exit      *string `json:"exit,omitempty"`
	Keepalive *int    `json:"keepalive,omitempty"`
	MTU       *int    `json:"mtu,omitempty"`
}

// Validate reports every bad field, and a request that changes nothing.
func (r *PrefsRequest) Validate() error {
	if r.Exit == nil && r.Keepalive == nil && r.MTU == nil {
		return errors.New("prefs: nothing to change; set exit, keepalive or mtu")
	}
	var errs []error
	if r.Exit != nil && *r.Exit != "" {
		if _, err := validate.Name("exit", *r.Exit); err != nil {
			errs = append(errs, err)
		}
	}
	if r.Keepalive != nil {
		if _, err := validate.IntRange("keepalive", *r.Keepalive, 0, 3600); err != nil {
			errs = append(errs, err)
		}
	}
	if r.MTU != nil {
		if _, err := validate.IntRange("mtu", *r.MTU, MinMTU, MaxMTU); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// PrefsResponse names the config that carries the change. Version is the
// config version the change created, 0 when no node's config changed;
// ConfigHash is the node's config now, which the agent has applied once its
// AppliedHash matches.
type PrefsResponse struct {
	Version       int64  `json:"version"`
	ConfigVersion int64  `json:"config_version"`
	ConfigHash    string `json:"config_hash"`
}

// DegradableSteps are the apply steps whose failure leaves the tunnel up;
// the mesh firewall degrades only when it takes its table away (amendment
// A43), exit mode only when it cannot hold a host's IPv6 back (A45).
var DegradableSteps = []string{"firewall", "containers", "dns", "policy", "exit"}

// The states of a node's split DNS (spec S7.4). Each one means the resolver
// serves the mesh zone on the node's mesh IP; they differ in whether the
// host's own lookups of .mesh names reach it.
const (
	// DNSOn: the host's resolver (systemd-resolved) sends .mesh to it.
	DNSOn = "on"
	// DNSOff: turned off on the node with `meshctl dns off`.
	DNSOff = "off"
	// DNSUnavailable: there is nothing to register with (no
	// systemd-resolved, no system bus, or the agent runs in a network
	// namespace the host's resolver cannot see).
	DNSUnavailable = "unavailable"
)

// DNSStates are the values of StatusReport.DNS.
var DNSStates = []string{DNSOn, DNSOff, DNSUnavailable}

// StatusResponse acknowledges a heartbeat. A successful one is also what
// confirms a safe-apply on server nodes.
type StatusResponse struct {
	ConfigVersion int64  `json:"config_version"`
	ConfigHash    string `json:"config_hash"`
	ServerTime    int64  `json:"server_time"`
}
