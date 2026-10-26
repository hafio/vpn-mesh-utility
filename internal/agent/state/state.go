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
	// JoinFile is the fingerprint of the join token `meshctl start` last
	// used (amendment A47), so a container joins once per token.
	JoinFile = "join.json"
	// LockFile is held by the one process that uses the data dir (Lock).
	LockFile = "agent.lock"
	// HostFile is the machine the data dir last ran on, as hostid.Local
	// (amendment A50): the cached config is applied at boot only there.
	HostFile = "host.json"
	// SessionFile is the agent's session with the control plane and the
	// ticket its next request presents (amendment A50), so a restart
	// carries on at once.
	SessionFile = "session.json"
	// RefusedFile says why the control plane refused this copy of the node
	// (amendment A50); it outlives the wipe after a foreign refusal.
	RefusedFile = "refused.json"
	// sigSuffix names the signature beside a cached body: config.json.sig.
	sigSuffix = ".sig"
)

// ErrNotJoined means the data directory has no enrollment yet.
var ErrNotJoined = errors.New("this node has not joined a mesh")

// Dir is an agent data directory.
type Dir struct {
	Path string
}

// Open returns the data directory at path, creating it (mode 0700) if
// needed. On Windows, where modes mean nothing, it gets an ACL instead
// (protect).
func Open(path string) (*Dir, error) {
	if _, err := validate.FilePath("--data", path); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w (run as root, or pass --data)", path, err)
	}
	if err := protect(path); err != nil {
		return nil, err
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
	// Mesh, MeshUUID and Slot are the node's mesh (amendment A49); the
	// slot names what the agent makes on the host.
	Mesh     string     `json:"mesh,omitempty"`
	MeshUUID string     `json:"mesh_uuid,omitempty"`
	Slot     proto.Slot `json:"slot,omitempty"`
	// ConfigKey is the control plane's public key pinned at join
	// (amendment A50): the agent trusts no config, mesh view or refusal
	// that it did not sign.
	ConfigKey string `json:"config_key"`
}

// Pinned is what the node checks its control plane's answers with.
func (e Enrollment) Pinned() proto.Pinned {
	key, _ := proto.ParseConfigKey(e.ConfigKey) // checked by Enrollment
	return proto.Pinned{Key: key, NodeID: e.NodeID, Slot: e.Slot}
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
	if _, err := proto.ParseConfigKey(e.ConfigKey); err != nil {
		return Enrollment{}, fmt.Errorf("enrollment %s has no valid control plane key (%w); re-join with a new token", d.file(EnrollmentFile), err)
	}
	if _, err := validate.ServerURL("server", e.Server); err != nil || e.NodeID <= 0 || !e.Slot.Valid() {
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

// NewKeys makes a fresh pair of keys. A join sends their public halves and
// writes them (SaveKeys) only once the control plane has accepted them
// (amendment A50), so a failed join leaves the old keys as they were.
func NewKeys() Keys {
	k := Keys{Identity: ed25519.NewKeyFromSeed(newSeed())}
	copy(k.WG[:], newWGKey())
	return k
}

// SaveKeys replaces both keys, each file atomically.
func (d *Dir) SaveKeys(k Keys) error {
	for _, f := range []struct {
		name string
		key  []byte
	}{
		{IdentityKeyFile, k.Identity.Seed()},
		{WGKeyFile, k.WG[:]},
	} {
		line := base64.StdEncoding.EncodeToString(f.key) + "\n"
		if err := fsutil.WriteAtomic(d.Path, f.name, []byte(line), 0o600); err != nil {
			return err
		}
	}
	return nil
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

// The caches hold the control plane's answers exactly as it signed them
// (amendment A50): the body in its file, the request nonce and signature in
// the file beside it (config.json.sig). Each is checked with the pinned key
// when it is read, so a cache changed on disk is ignored like a forged
// answer.

// sigFile is the signature beside a cached body.
type sigFile struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	// FetchedAt is when a mesh view was fetched (unix seconds).
	FetchedAt int64 `json:"fetched_at,omitempty"`
}

// writeSigned caches b in name, the signature first: a crash in between
// leaves a pair that does not verify, which is ignored.
func (d *Dir) writeSigned(name string, b proto.SignedBody, fetchedAt int64) error {
	sig, err := json.Marshal(sigFile{Nonce: b.Nonce, Signature: b.Signature, FetchedAt: fetchedAt})
	if err != nil {
		return fmt.Errorf("encode %s%s: %w", name, sigSuffix, err)
	}
	if err := fsutil.WriteAtomic(d.Path, name+sigSuffix, append(sig, '\n'), 0o600); err != nil {
		return err
	}
	return fsutil.WriteAtomic(d.Path, name, b.Body, 0o600)
}

// readSigned reads a cached body and its signature, checked with p as kind;
// nothing cached is nil.
func (d *Dir) readSigned(name, kind string, p proto.Pinned) (*proto.SignedBody, int64, error) {
	body, err := os.ReadFile(d.file(name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, fmt.Errorf("read %s: %w", d.file(name), err)
	}
	raw, err := os.ReadFile(d.file(name + sigSuffix))
	var sig sigFile
	if err == nil {
		err = json.Unmarshal(raw, &sig)
	}
	b := proto.SignedBody{Body: body, Nonce: sig.Nonce, Signature: sig.Signature}
	if err == nil {
		err = p.VerifyBody(kind, b)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("%s is not as the control plane signed it (%w); it is ignored until the server sends a new one", d.file(name), err)
	}
	return &b, sig.FetchedAt, nil
}

// removeSigned drops a cached body and its signature.
func (d *Dir) removeSigned(name string) error {
	for _, f := range []string{name, name + sigSuffix} {
		if err := os.Remove(d.file(f)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", d.file(f), err)
		}
	}
	return nil
}

// Config returns the cached config, checked with p, or nil when there is
// none.
func (d *Dir) Config(p proto.Pinned) (*proto.NodeConfig, error) { return d.readConfig(ConfigFile, p) }

// PendingConfig returns the config safe apply was waiting to confirm,
// checked with p, or nil.
func (d *Dir) PendingConfig(p proto.Pinned) (*proto.NodeConfig, error) {
	return d.readConfig(PendingConfigFile, p)
}

func (d *Dir) readConfig(name string, p proto.Pinned) (*proto.NodeConfig, error) {
	b, _, err := d.readSigned(name, proto.SignedConfig, p)
	if b == nil || err != nil {
		return nil, err
	}
	var c proto.NodeConfig
	if err := json.Unmarshal(b.Body, &c); err != nil {
		return nil, fmt.Errorf("cached config %s is damaged (%w); it is ignored until the server sends a new one", d.file(name), err)
	}
	return &c, nil
}

// SavePending records a config as applied but not yet confirmed.
func (d *Dir) SavePending(b proto.SignedBody) error { return d.writeSigned(PendingConfigFile, b, 0) }

// ClearPending drops the pending config: it was confirmed or rolled back.
func (d *Dir) ClearPending() error { return d.removeSigned(PendingConfigFile) }

// SaveConfig caches a config as the one to re-apply at boot, keeping the one
// it replaces as config.prev.json.
func (d *Dir) SaveConfig(b proto.SignedBody) error {
	for _, f := range []string{ConfigFile + sigSuffix, ConfigFile} {
		prev, err := os.ReadFile(d.file(f))
		if err != nil {
			continue
		}
		if err := fsutil.WriteAtomic(d.Path, strings.Replace(f, ConfigFile, PrevConfigFile, 1), prev, 0o600); err != nil {
			return err
		}
	}
	return d.writeSigned(ConfigFile, b, 0)
}

// ClearConfig drops the cached configs; join calls it when the node now
// belongs to another mesh or another node id.
func (d *Dir) ClearConfig() error {
	for _, f := range []string{ConfigFile, PrevConfigFile, PendingConfigFile} {
		if err := d.removeSigned(f); err != nil {
			return err
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

// Mesh reads the cached mesh view, checked with p; none yet is the zero
// value.
func (d *Dir) Mesh(p proto.Pinned) (proto.LocalMesh, error) {
	b, at, err := d.readSigned(MeshFile, proto.SignedMesh, p)
	if b == nil || err != nil {
		return proto.LocalMesh{}, err
	}
	m := proto.LocalMesh{FetchedAt: at}
	if err = json.Unmarshal(b.Body, &m.View); err == nil {
		err = m.View.Validate()
	}
	if err != nil {
		return proto.LocalMesh{}, fmt.Errorf("cached mesh view %s is damaged (%w); it is ignored until the server sends a new one", d.file(MeshFile), err)
	}
	return m, nil
}

// SaveMesh caches a mesh view the server signed, fetched at fetchedAt.
func (d *Dir) SaveMesh(b proto.SignedBody, fetchedAt int64) error {
	return d.writeSigned(MeshFile, b, fetchedAt)
}

// JoinRecord is what `meshctl start` remembers of the last join token it
// used, or that the server refused: a fingerprint, never the token.
type JoinRecord struct {
	Token string `json:"token_fingerprint"`
}

// JoinRecord reads join.json; none yet is the zero value.
func (d *Dir) JoinRecord() (JoinRecord, error) {
	b, err := os.ReadFile(d.file(JoinFile))
	if errors.Is(err, fs.ErrNotExist) {
		return JoinRecord{}, nil
	}
	if err != nil {
		return JoinRecord{}, fmt.Errorf("read %s: %w", d.file(JoinFile), err)
	}
	var r JoinRecord
	if err := json.Unmarshal(b, &r); err != nil {
		return JoinRecord{}, fmt.Errorf("%s is damaged (%w); remove it, and the next start joins with the token it is given", d.file(JoinFile), err)
	}
	return r, nil
}

// SaveJoinRecord writes join.json.
func (d *Dir) SaveJoinRecord(r JoinRecord) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", JoinFile, err)
	}
	return fsutil.WriteAtomic(d.Path, JoinFile, append(b, '\n'), 0o600)
}
