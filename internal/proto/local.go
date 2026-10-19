package proto

// The agent's local API: what `meshctl status`, `meshctl exit`,
// `meshctl dns` and `meshctl logs` (and the local UI) use over the agent
// socket. It never leaves the node.

// Local API paths. LocalPathPrefs takes a PrefsRequest, which the agent
// relays to the server as a signed request, and answers its PrefsResponse.
// LocalPathDNS takes a DNSRequest, which stays on the node (amendment A10),
// and answers the LocalStatus after it. LocalPathMesh answers a LocalMesh;
// LocalPathLogs the agent's recent log lines as text, and with
// ?follow=true keeps sending new ones. LocalPathUIToken answers a UIToken.
const (
	LocalPathStatus  = "/v1/status"
	LocalPathPrefs   = "/v1/prefs"
	LocalPathDNS     = "/v1/dns"
	LocalPathMesh    = "/v1/mesh"
	LocalPathLogs    = "/v1/logs"
	LocalPathUIToken = "/v1/ui-token"
)

// UIToken is a one-time sign-in link to the local UI (amendment A7), for
// `meshctl ui`: whoever can open the agent socket may sign in.
type UIToken struct {
	URL string `json:"url"`
	// ExpiresAt is when the link stops working if unused (unix seconds).
	ExpiresAt int64 `json:"expires_at"`
}

// LocalMesh is the mesh view as this node has it: fetched from the server
// when it answers, else the one cached in mesh.json, so the view stays
// readable while the server is down (spec S7.6).
type LocalMesh struct {
	View MeshView `json:"view"`
	// FetchedAt is when the agent got View from the server (unix
	// seconds); zero means it never has.
	FetchedAt int64 `json:"fetched_at,omitempty"`
	// Error says why View is not fresh: the server could not be asked.
	Error string `json:"error,omitempty"`
}

// DNSRequest turns the node's split DNS on or off.
type DNSRequest struct {
	On bool `json:"on"`
}

// LocalStatus is the agent's view of itself and its peers.
type LocalStatus struct {
	Node         Self   `json:"node"`
	MeshIP       string `json:"mesh_ip"`
	Server       string `json:"server"`
	AgentVersion string `json:"agent_version"`
	WGBackend    string `json:"wg_backend"`
	// AppliedVersion and AppliedHash identify the config on the interface;
	// zero until the first successful apply.
	AppliedVersion int64  `json:"applied_version"`
	AppliedHash    string `json:"applied_hash,omitempty"`
	AppliedAt      int64  `json:"applied_at,omitempty"`
	// LastContact is the last time the server answered (unix seconds).
	LastContact int64        `json:"last_contact,omitempty"`
	Peers       []PeerStatus `json:"peers"`
	Errors      []string     `json:"errors,omitempty"`
	Degraded    []string     `json:"degraded,omitempty"`
	Exit        string       `json:"exit,omitempty"`
	// ExitChoices are the servers this node may choose as its exit.
	ExitChoices []string `json:"exit_choices,omitempty"`
	// Rollback is set while safe apply has the node on an older config
	// than the server's (spec S6.4).
	Rollback *Rollback `json:"rollback,omitempty"`
	// DNS is the node's split DNS (DNSStates); DNSOff is kept across
	// restarts until `meshctl dns on`. Zone is the zone it serves.
	DNS  string `json:"dns,omitempty"`
	Zone string `json:"zone,omitempty"`
	// MTU and Keepalive are the applied config's (Keepalive: a spoke's, in
	// seconds); zero before the first apply.
	MTU       int `json:"mtu,omitempty"`
	Keepalive int `json:"keepalive,omitempty"`
}
