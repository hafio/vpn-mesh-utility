package proto

import (
	"errors"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Container names (amendment A38): an agent whose containers are on the
// mesh reports the names its runtime gives them, and the server publishes
// them in the mesh view, so every node's resolver answers them. They stay
// out of NodeConfig: a container restart is not a config version.

// PathNames is where an agent puts its containers' names (PUT, signed).
const PathNames = "/api/v1/agent/names"

// EventMesh is the SSE event that tells an agent the mesh view changed
// (container names, for now); its data is an empty JSON object.
const EventMesh = "mesh"

// Where a container's name comes from.
const (
	// NameService is a compose service (com.docker.compose.service).
	NameService = "service"
	// NameSwarm is a Swarm service without its stack's prefix.
	NameSwarm = "swarm"
	// NameContainer is the container's own name.
	NameContainer = "container"
	// NameAlias is a network alias on the mesh network.
	NameAlias = "alias"
	// NameHostname is the container's hostname, or hostname.domainname.
	NameHostname = "hostname"
)

// NameKinds are the values of ContainerName.Kind.
var NameKinds = []string{NameService, NameSwarm, NameContainer, NameAlias, NameHostname}

// MaxNames caps the names one node reports.
const MaxNames = 512

// ContainerName is one name of one container on the node's mesh network.
// Name is lower case: one label ("web"), or a dotted name the container
// was given ("web.example.home.arpa"). IP is of either family: a container
// with IPv6 on the mesh (M8) has its name once per address.
type ContainerName struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	IP   string `json:"ip"`
}

// NamesRequest is every name of the node's containers on the mesh network:
// it replaces what the server held, and an empty list clears it.
type NamesRequest struct {
	Names []ContainerName `json:"names"`
}

// Validate normalises the request in place and reports every bad entry,
// and a name given twice for one address.
func (r *NamesRequest) Validate() error {
	if len(r.Names) > MaxNames {
		return fmt.Errorf("names: at most %d entries (got %d)", MaxNames, len(r.Names))
	}
	var errs []error
	seen := make(map[ContainerName]bool, len(r.Names))
	for i := range r.Names {
		n := &r.Names[i]
		field := fmt.Sprintf("names[%d]", i)
		name, err := validate.Hostname(field+".name", n.Name)
		if err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.Enum(field+".kind", n.Kind, NameKinds...); err != nil {
			errs = append(errs, err)
		}
		ip, err := validate.IP(field+".ip", n.IP)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		n.Name, n.IP = name, ip.String()
		key := ContainerName{Name: n.Name, IP: n.IP}
		if seen[key] {
			errs = append(errs, fmt.Errorf("%s: %s -> %s is listed twice", field, n.Name, n.IP))
		}
		seen[key] = true
	}
	return errors.Join(errs...)
}

// NamesResponse says what the server did with the names: how many it
// keeps, and the ones it does not publish, with why.
type NamesResponse struct {
	Stored  int           `json:"stored"`
	Dropped []DroppedName `json:"dropped,omitempty"`
}

// DroppedName is a name the server does not publish.
type DroppedName struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

// MeshName is one published container name: Name is a label of the mesh
// zone, answered as <name>.<zone> (every node's addresses, round robin)
// and <name>.<node>.<zone> (this node's).
type MeshName struct {
	Name string `json:"name"`
	Node string `json:"node"`
	IP   string `json:"ip"`
}

// maxMeshNames caps the names of a mesh view: MaxNames for each of the 128
// container networks a /16 mesh holds.
const maxMeshNames = 1 << 16

func validateMeshNames(names []MeshName) []error {
	if len(names) > maxMeshNames {
		return []error{fmt.Errorf("names: at most %d entries (got %d)", maxMeshNames, len(names))}
	}
	var errs []error
	for i, n := range names {
		field := fmt.Sprintf("names[%d]", i)
		if _, err := validate.DNSName(field+".name", n.Name); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.Name(field+".node", n.Node); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.IP(field+".ip", n.IP); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
