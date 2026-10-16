// Package state is the agent's data directory (spec S6.1, S6.3): its two
// private keys, what join learned, the cached config the agent applies at
// boot without asking the server, the node's own settings, and the mesh
// view the local UI shows while the server is down.
//
// Private keys are generated here and never leave the node. Every write goes
// through fsutil.WriteAtomic, so a crash or a power cut leaves the old file or
// the new one, never half of either.
package state

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Files in the data directory.
const (
	IdentityKeyFile = "identity.key"
	WGKeyFile       = "wg.key"
	EnrollmentFile  = "state.json"
	ConfigFile      = "config.json"
	PrevConfigFile  = "config.prev.json"
	// PendingConfigFile is a server node's config that is applied but not
	// yet confirmed by safe apply (spec S6.4); config.json stays the last
	// confirmed one until it is.
	PendingConfigFile = "config.pending.json"
	// LocalFile holds the node's own settings (`meshctl dns off`), which
	// never go to the server (amendment A10).
	LocalFile = "local.json"
	// MeshFile is the last mesh view fetched from the server (spec S7.6).
	MeshFile = "mesh.json"
)

// ErrNotJoined means the data directory has no enrollment yet.
var ErrNotJoined = errors.New("this node has not joined a mesh")

// Dir is an agent data directory.
type Dir struct {
	Path string
}

// Open returns the data directory at path, creating it (mode 0700) if needed.
func Open(path string) (*Dir, error) {
	if _, err := validate.FilePath("--data", path); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w (run as root, or pass --data)", path, err)
	}
	return &Dir{Path: path}, nil
}

func (d *Dir) file(name string) string { return filepath.Join(d.Path, name) }

// Enrollment is what the server told the agent at join.
type Enrollment struct {
	Server   string     `json:"server"`
	NodeID   int64      `json:"node_id"`
	Name     string     `json:"name"`
	Role     proto.Role `json:"role"`
	MeshIP   string     `json:"mesh_ip"`
	JoinedAt int64      `json:"joined_at"`
}

// Enrollment reads state.json; ErrNotJoined when there is none.
func (d *Dir) Enrollment() (Enrollment, error) {
	b, err := os.ReadFile(d.file(EnrollmentFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Enrollment{}, ErrNotJoined
	}
	if err != nil {
		return Enrollment{}, fmt.Errorf("read enrollment: %w", err)
	}
	var e Enrollment
	if err := json.Unmarshal(b, &e); err != nil {
		return Enrollment{}, fmt.Errorf("read enrollment %s: %w (re-join to rewrite it)", d.file(EnrollmentFile), err)
	}
	if _, err := validate.ServerURL("server", e.Server); err != nil || e.NodeID <= 0 {
		return Enrollment{}, fmt.Errorf("enrollment %s is damaged; re-join to rewrite it", d.file(EnrollmentFile))
	}
	return e, nil
}

// SaveEnrollment writes state.json.
func (d *Dir) SaveEnrollment(e Enrollment) error {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("encode enrollment: %w", err)
	}
	return fsutil.WriteAtomic(d.Path, EnrollmentFile, append(b, '\n'), 0o600)
}

// Keys are the node's private keys.
type Keys struct {
	// Identity signs every API request (spec S6.2).
	Identity ed25519.PrivateKey
	// WG is the WireGuard (X25519) private key, clamped.
	WG [32]byte
}

// IdentityPublic is the base64 identity public key sent at join.
func (k Keys) IdentityPublic() string {
	return base64.StdEncoding.EncodeToString(k.Identity.Public().(ed25519.PublicKey))
}

// WGPublic is the base64 WireGuard public key sent at join.
func (k Keys) WGPublic() (string, error) {
	priv, err := ecdh.X25519().NewPrivateKey(k.WG[:])
	if err != nil {
		return "", fmt.Errorf("derive WireGuard public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes()), nil
}

// LoadKeys reads both keys; a missing key means the node never joined.
func (d *Dir) LoadKeys() (Keys, error) {
	seed, err := d.readKey(IdentityKeyFile, ed25519.SeedSize)
	if err != nil {
		return Keys{}, err
	}
	wg, err := d.readKey(WGKeyFile, 32)
	if err != nil {
		return Keys{}, err
	}
	k := Keys{Identity: ed25519.NewKeyFromSeed(seed)}
	copy(k.WG[:], wg)
	return k, nil
}

// LoadOrCreateKeys reads both keys, generating whichever is missing. Re-joining
// keeps existing keys, so a node that only needed a new token keeps its
// identity.
func (d *Dir) LoadOrCreateKeys() (Keys, error) {
	for _, f := range []struct {
		name string
		gen  func() []byte
	}{
		{IdentityKeyFile, newSeed},
		{WGKeyFile, newWGKey},
	} {
		_, err := os.Stat(d.file(f.name))
		if err == nil {
			continue
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return Keys{}, fmt.Errorf("check %s: %w", d.file(f.name), err)
		}
		line := base64.StdEncoding.EncodeToString(f.gen()) + "\n"
		if err := fsutil.WriteAtomic(d.Path, f.name, []byte(line), 0o600); err != nil {
			return Keys{}, err
		}
	}
	return d.LoadKeys()
}

func newSeed() []byte {
	b := make([]byte, ed25519.SeedSize)
	_, _ = rand.Read(b) // crypto/rand.Read never fails (Go 1.24+)
	return b
}

// newWGKey is what `wg genkey` makes: 32 random bytes, clamped for X25519.
func newWGKey() []byte {
	b := make([]byte, 32)
	_, _ = rand.Read(b) // crypto/rand.Read never fails (Go 1.24+)
	b[0] &= 248
	b[31] = (b[31] & 127) | 64
	return b
}

func (d *Dir) readKey(name string, size int) ([]byte, error) {
	path := d.file(name)
	st, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s is missing: %w; run `meshctl agent join` first", path, ErrNotJoined)
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	// File modes mean little on Windows; its ACLs are set at install (M5).
	if runtime.GOOS != "windows" && st.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%s is readable by other users (mode %04o); run `chmod 600 %s`", path, st.Mode().Perm(), path)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	key, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(string(b)))
	if err != nil || len(key) != size {
		return nil, fmt.Errorf("%s does not hold a base64 %d-byte key; restore it from a backup or re-join after removing it", path, size)
	}
	return key, nil
}

// Config returns the cached config, or nil when there is none.
func (d *Dir) Config() (*proto.NodeConfig, error) { return d.readConfig(ConfigFile) }

// PendingConfig returns the config safe apply was waiting to confirm, or nil.
func (d *Dir) PendingConfig() (*proto.NodeConfig, error) { return d.readConfig(PendingConfigFile) }

func (d *Dir) readConfig(name string) (*proto.NodeConfig, error) {
	b, err := os.ReadFile(d.file(name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", d.file(name), err)
	}
	var c proto.NodeConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("cached config %s is damaged (%w); it is ignored until the server sends a new one", d.file(name), err)
	}
	return &c, nil
}

// SavePending records c as applied but not yet confirmed.
func (d *Dir) SavePending(c proto.NodeConfig) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	return fsutil.WriteAtomic(d.Path, PendingConfigFile, append(b, '\n'), 0o600)
}

// ClearPending drops the pending config: it was confirmed or rolled back.
func (d *Dir) ClearPending() error {
	if err := os.Remove(d.file(PendingConfigFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove %s: %w", d.file(PendingConfigFile), err)
	}
	return nil
}

// SaveConfig caches c as the config to re-apply at boot, keeping the one it
// replaces as config.prev.json.
func (d *Dir) SaveConfig(c proto.NodeConfig) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if prev, err := os.ReadFile(d.file(ConfigFile)); err == nil {
		if err := fsutil.WriteAtomic(d.Path, PrevConfigFile, prev, 0o600); err != nil {
			return err
		}
	}
	return fsutil.WriteAtomic(d.Path, ConfigFile, append(b, '\n'), 0o600)
}

// ClearConfig drops the cached configs; join calls it when the node now
// belongs to another mesh or another node id.
func (d *Dir) ClearConfig() error {
	for _, f := range []string{ConfigFile, PrevConfigFile, PendingConfigFile} {
		if err := os.Remove(d.file(f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", d.file(f), err)
		}
	}
	return nil
}

// Local is the node's own settings.
type Local struct {
	DNSOff bool `json:"dns_off,omitempty"`
}

// Local reads the node's own settings; none saved yet is the zero value.
func (d *Dir) Local() (Local, error) {
	b, err := os.ReadFile(d.file(LocalFile))
	if errors.Is(err, fs.ErrNotExist) {
		return Local{}, nil
	}
	if err != nil {
		return Local{}, fmt.Errorf("read %s: %w", d.file(LocalFile), err)
	}
	var l Local
	if err := json.Unmarshal(b, &l); err != nil {
		return Local{}, fmt.Errorf("%s is damaged (%w); remove it to go back to the defaults (DNS on)", d.file(LocalFile), err)
	}
	return l, nil
}

// SaveLocal writes the node's own settings.
func (d *Dir) SaveLocal(l Local) error {
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", LocalFile, err)
	}
	return fsutil.WriteAtomic(d.Path, LocalFile, append(b, '\n'), 0o600)
}

// Mesh reads the cached mesh view; none yet is the zero value.
func (d *Dir) Mesh() (proto.LocalMesh, error) {
	b, err := os.ReadFile(d.file(MeshFile))
	if errors.Is(err, fs.ErrNotExist) {
		return proto.LocalMesh{}, nil
	}
	if err != nil {
		return proto.LocalMesh{}, fmt.Errorf("read %s: %w", d.file(MeshFile), err)
	}
	var m proto.LocalMesh
	if err = json.Unmarshal(b, &m); err == nil {
		err = m.View.Validate()
	}
	if err != nil {
		return proto.LocalMesh{}, fmt.Errorf("cached mesh view %s is damaged (%w); it is ignored until the server sends a new one", d.file(MeshFile), err)
	}
	m.Error = ""
	return m, nil
}

// SaveMesh caches a mesh view fetched from the server.
func (d *Dir) SaveMesh(m proto.LocalMesh) error {
	m.Error = ""
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", MeshFile, err)
	}
	return fsutil.WriteAtomic(d.Path, MeshFile, append(b, '\n'), 0o600)
}
