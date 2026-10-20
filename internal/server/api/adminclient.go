package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// AdminClient is the `meshctl node` side of the admin socket.
type AdminClient struct {
	socket string
	http   *http.Client
}

// NewAdminClient returns a client for the server listening on socket.
func NewAdminClient(socket string) *AdminClient {
	tr := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socket)
		},
	}
	return &AdminClient{socket: socket, http: &http.Client{Transport: tr, Timeout: 30 * time.Second}}
}

// AdminError is a refusal from the server, with its machine-readable code.
type AdminError struct {
	Status  int
	Code    string
	Message string
}

func (e *AdminError) Error() string { return e.Message }

func (c *AdminClient) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, "http://meshctl"+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("reach the server at %s: %w (is `meshctl server` running with this --data, and are you its user?)", c.socket, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return fmt.Errorf("read the server's answer: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		var eb proto.ErrorBody
		if json.Unmarshal(b, &eb) == nil && eb.Message != "" {
			return &AdminError{Status: resp.StatusCode, Code: eb.Code, Message: eb.Message}
		}
		return fmt.Errorf("the server answered %s", resp.Status)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("decode the server's answer: %w", err)
	}
	return nil
}

func nodePath(name, suffix string) string { return AdminNodes + "/" + url.PathEscape(name) + suffix }

// List returns every node and the plan's problems.
func (c *AdminClient) List(ctx context.Context) (NodeList, error) {
	var out NodeList
	err := c.do(ctx, http.MethodGet, AdminNodes, nil, &out)
	return out, err
}

// Add creates a node and its first join token.
func (c *AdminClient) Add(ctx context.Context, req NodeAddRequest) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodPost, AdminNodes, req, &out)
	return out, err
}

// Show returns one node with its computed config.
func (c *AdminClient) Show(ctx context.Context, name string) (NodeDetail, error) {
	var out NodeDetail
	err := c.do(ctx, http.MethodGet, nodePath(name, ""), nil, &out)
	return out, err
}

// Set changes the given fields of a node.
func (c *AdminClient) Set(ctx context.Context, name string, req NodeSetRequest) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodPatch, nodePath(name, ""), req, &out)
	return out, err
}

// Disable takes a node out of every peer list.
func (c *AdminClient) Disable(ctx context.Context, name string) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodPost, nodePath(name, "/disable"), nil, &out)
	return out, err
}

// Enable puts a disabled node back.
func (c *AdminClient) Enable(ctx context.Context, name string) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodPost, nodePath(name, "/enable"), nil, &out)
	return out, err
}

// Remove deletes a node.
func (c *AdminClient) Remove(ctx context.Context, name string) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodDelete, nodePath(name, ""), nil, &out)
	return out, err
}

// Compose returns the compose file that runs a node's agent in a
// container; empty options take the server's defaults for its mode.
func (c *AdminClient) Compose(ctx context.Context, name string, o ComposeOptions) (ComposeFile, error) {
	q := url.Values{}
	for k, v := range map[string]string{"backend": o.Backend, "data": o.DataDir, "runtime": o.Runtime} {
		if v != "" {
			q.Set(k, v)
		}
	}
	if o.NoNames {
		q.Set("names", "off")
	}
	path := nodePath(name, "/compose")
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out ComposeFile
	err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

// Network returns the commands that create a node's mesh network for
// runtime ("" is docker).
func (c *AdminClient) Network(ctx context.Context, name, runtime string) (NetworkInfo, error) {
	path := nodePath(name, "/network")
	if runtime != "" {
		path += "?" + url.Values{"runtime": {runtime}}.Encode()
	}
	var out NetworkInfo
	err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

// Token issues a new join token, revoking unused ones.
func (c *AdminClient) Token(ctx context.Context, name string) (NodeChange, error) {
	var out NodeChange
	err := c.do(ctx, http.MethodPost, nodePath(name, "/token"), nil, &out)
	return out, err
}

// Phone makes a phone's profile with a new key pair; an empty exit is the
// mesh-only profile.
func (c *AdminClient) Phone(ctx context.Context, name, exit string) (PhoneProfile, error) {
	var out PhoneProfile
	err := c.do(ctx, http.MethodPost, nodePath(name, "/phone"), PhoneRequest{Exit: exit}, &out)
	return out, err
}
