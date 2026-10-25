package names

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Sockets are where the agent looks for a runtime, in order: Docker's,
// then rootful Podman's. In container mode the compose file mounts the
// host's at the first.
func Sockets() []string { return []string{"/var/run/docker.sock", "/run/podman/podman.sock"} }

// Limits on what is read from the runtime.
const (
	callTimeout  = 10 * time.Second
	maxList      = 8 << 20
	maxInspect   = 1 << 20
	maxEventLine = 64 << 10
	// maxContainers caps the containers inspected per listing.
	maxContainers = 1024
)

// reID is a container ID as Docker and Podman give it.
var reID = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Source is where the agent reads the containers from: a Runtime, or a
// test's fake.
type Source interface {
	// List returns the running containers on the mesh network.
	List(ctx context.Context) ([]Container, error)
	// Watch calls changed whenever a container or network changes, until
	// it fails or ctx is cancelled.
	Watch(ctx context.Context, changed func()) error
}

// Runtime is a container runtime's Docker-compatible API on a unix socket.
type Runtime struct {
	socket  string
	network string
	calls   *http.Client
	streams *http.Client
}

// New returns the runtime answering on socket, whose containers on network
// (the agent's mesh's, proto.Slot.Network) are the mesh's.
func New(socket, network string) *Runtime {
	tr := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socket)
		},
		MaxIdleConns:    2,
		IdleConnTimeout: time.Minute,
	}
	return &Runtime{socket: socket, network: network, calls: &http.Client{Transport: tr, Timeout: callTimeout}, streams: &http.Client{Transport: tr}}
}

// Detect returns the first of paths that is a socket, "" when none is.
func Detect(paths []string) string {
	for _, s := range paths {
		if fi, err := os.Stat(s); err == nil && fi.Mode().Type() == fs.ModeSocket {
			return s
		}
	}
	return ""
}

// Socket is where the runtime answers.
func (r *Runtime) Socket() string { return r.socket }

func (r *Runtime) get(ctx context.Context, c *http.Client, path string, query url.Values) (*http.Response, error) {
	u := url.URL{Scheme: "http", Host: "localhost", Path: path, RawQuery: query.Encode()}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("container runtime on %s: GET %s: %w", r.socket, path, err)
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512)) // best effort: the status already failed
		_ = resp.Body.Close()
		return nil, &StatusError{Path: path, Status: resp.StatusCode, Message: clip(string(b))}
	}
	return resp, nil
}

// StatusError is a runtime's refusal.
type StatusError struct {
	Path    string
	Status  int
	Message string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("container runtime: GET %s answered %d: %s", e.Path, e.Status, e.Message)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 120 {
		return s[:120] + "..."
	}
	return s
}

func (r *Runtime) getJSON(ctx context.Context, path string, query url.Values, max int64, out any) error {
	resp, err := r.get(ctx, r.calls, path, query)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return fmt.Errorf("container runtime: read GET %s: %w", path, err)
	}
	if int64(len(b)) > max {
		return fmt.Errorf("container runtime: GET %s answered more than %d bytes", path, max)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("container runtime: decode GET %s: %w", path, err)
	}
	return nil
}

// meshFilter selects the containers, or events, of the mesh network.
func meshFilter(key string, values ...string) url.Values {
	f, _ := json.Marshal(map[string][]string{key: values}) // cannot fail
	return url.Values{"filters": {string(f)}}
}

type listed struct {
	ID string `json:"Id"`
}

type inspected struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Config struct {
		Hostname   string            `json:"Hostname"`
		Domainname string            `json:"Domainname"`
		Labels     map[string]string `json:"Labels"`
	} `json:"Config"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress         string   `json:"IPAddress"`
			GlobalIPv6Address string   `json:"GlobalIPv6Address"`
			Aliases           []string `json:"Aliases"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

// List returns the running containers on the mesh network.
func (r *Runtime) List(ctx context.Context) ([]Container, error) {
	var ls []listed
	if err := r.getJSON(ctx, "/containers/json", meshFilter("network", r.network), maxList, &ls); err != nil {
		return nil, err
	}
	if len(ls) > maxContainers {
		ls = ls[:maxContainers]
	}
	var out []Container
	for _, l := range ls {
		if !reID.MatchString(l.ID) {
			continue // never part of a path
		}
		var in inspected
		err := r.getJSON(ctx, "/containers/"+l.ID+"/json", nil, maxInspect, &in)
		var se *StatusError
		if errors.As(err, &se) && se.Status == http.StatusNotFound {
			continue // gone since the listing
		}
		if err != nil {
			return nil, err
		}
		n, ok := in.NetworkSettings.Networks[r.network]
		ip, _ := netip.ParseAddr(n.IPAddress)
		ip6, _ := netip.ParseAddr(n.GlobalIPv6Address) // M8; "" without IPv6
		if !ok || (!ip.IsValid() && !ip6.IsValid()) || in.ID != l.ID {
			continue
		}
		out = append(out, Container{ID: in.ID, Name: in.Name, Hostname: in.Config.Hostname, Domainname: in.Config.Domainname,
			Labels: in.Config.Labels, IP: ip, IP6: ip6, Aliases: n.Aliases})
	}
	return out, nil
}

// Watch calls changed for every container or network event until the
// stream ends or ctx is cancelled; it always returns an error saying why.
func (r *Runtime) Watch(ctx context.Context, changed func()) error {
	resp, err := r.get(ctx, r.streams, "/events", meshFilter("type", "container", "network"))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 4096), maxEventLine)
	for sc.Scan() {
		if strings.TrimSpace(sc.Text()) != "" {
			changed()
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("container runtime: event stream: %w", err)
	}
	return errors.New("container runtime: event stream ended")
}
