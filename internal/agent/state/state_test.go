package state

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func open(t *testing.T) *Dir {
	t.Helper()
	d, err := Open(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// Every join makes fresh keys (amendment A50); saved, they read back as
// they were, each file 0600.
func TestFreshKeysAreSavedAndRead(t *testing.T) {
	d := open(t)
	if _, err := d.LoadKeys(); !errors.Is(err, ErrNotJoined) {
		t.Fatalf("LoadKeys before join = %v, want ErrNotJoined", err)
	}
	k1, other := NewKeys(), NewKeys()
	if k1.Identity.Equal(other.Identity) || k1.WG == other.WG {
		t.Fatal("two joins made the same keys")
	}
	if err := d.SaveKeys(k1); err != nil {
		t.Fatal(err)
	}
	k2, err := d.LoadKeys()
	if err != nil {
		t.Fatal(err)
	}
	if !k1.Identity.Equal(k2.Identity) || k1.WG != k2.WG {
		t.Fatal("the saved keys read back differently")
	}
	if k1.WG[0]&7 != 0 || k1.WG[31]&128 != 0 || k1.WG[31]&64 == 0 {
		t.Fatal("the WireGuard key is not clamped")
	}
	pub, err := k1.WGPublic()
	if err != nil || len(pub) != 44 {
		t.Fatalf("WGPublic = %q, %v", pub, err)
	}
	idPub, _ := base64.StdEncoding.DecodeString(k1.IdentityPublic())
	if len(idPub) != ed25519.PublicKeySize {
		t.Fatalf("IdentityPublic = %q", k1.IdentityPublic())
	}
	if runtime.GOOS != "windows" {
		for _, f := range []string{IdentityKeyFile, WGKeyFile} {
			st, err := os.Stat(filepath.Join(d.Path, f))
			if err != nil || st.Mode().Perm() != 0o600 {
				t.Fatalf("%s mode %v, %v; want 0600", f, st.Mode().Perm(), err)
			}
		}
	}
}

func TestLoadKeysRefusesLooseOrDamagedFiles(t *testing.T) {
	d := open(t)
	if err := d.SaveKeys(NewKeys()); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filepath.Join(d.Path, WGKeyFile), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := d.LoadKeys(); err == nil || !strings.Contains(err.Error(), "chmod 600") {
			t.Fatalf("LoadKeys with a world-readable key = %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(d.Path, WGKeyFile), []byte("not base64!\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(d.Path, WGKeyFile), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := d.LoadKeys(); err == nil || !strings.Contains(err.Error(), "32-byte key") {
		t.Fatalf("LoadKeys with a damaged key = %v", err)
	}
}

func TestEnrollmentRoundTrip(t *testing.T) {
	d := open(t)
	if _, err := d.Enrollment(); !errors.Is(err, ErrNotJoined) {
		t.Fatalf("Enrollment before join = %v", err)
	}
	e := Enrollment{Server: "https://mesh.example.com", NodeID: 4, Name: "nas", Role: proto.RoleSpoke, MeshIP: "10.77.1.1", JoinedAt: testfix.Epoch.Unix(),
		ConfigKey: testfix.ConfigKeyB64()}
	if err := d.SaveEnrollment(e); err != nil {
		t.Fatal(err)
	}
	got, err := d.Enrollment()
	if err != nil || got != e {
		t.Fatalf("Enrollment = %+v, %v", got, err)
	}
	// The node's mesh and its slot (amendment A49).
	e.Mesh, e.MeshUUID, e.Slot = "lab", "8c2b6f3e-0d1a-4c55-9e7a-1f2d3c4b5a69", 3
	if err := d.SaveEnrollment(e); err != nil {
		t.Fatal(err)
	}
	if got, err := d.Enrollment(); err != nil || got != e {
		t.Fatalf("Enrollment in slot 3 = %+v, %v", got, err)
	}
	if p := e.Pinned(); p.NodeID != 4 || p.Slot != 3 || !p.Key.Equal(testfix.Pinned(4).Key) {
		t.Fatalf("Pinned = %+v", p)
	}
	key := `,"config_key":"` + testfix.ConfigKeyB64() + `"`
	for _, bad := range []string{`{"server":"ftp://x","node_id":4` + key + `}`, `{"server":"https://mesh.example.com","node_id":4,"slot":64` + key + `}`,
		`{"server":"https://mesh.example.com","node_id":4}`, `{"server":"https://mesh.example.com","node_id":4,"config_key":"c2hvcnQ="}`} {
		if err := os.WriteFile(filepath.Join(d.Path, EnrollmentFile), []byte(bad), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := d.Enrollment(); err == nil || !strings.Contains(err.Error(), "re-join") {
			t.Fatalf("damaged enrollment %s = %v", bad, err)
		}
	}
}

// vps1 is what the agent of vps-1 checks its caches with.
var vps1 = testfix.Pinned(testfix.VPS1.ID)

func TestConfigCacheKeepsThePrevious(t *testing.T) {
	d := open(t)
	if c, err := d.Config(vps1); c != nil || err != nil {
		t.Fatalf("Config before any = %v, %v", c, err)
	}
	a := testfix.ServerConfig()
	b := testfix.ServerConfig()
	b.Version = 4
	if err := d.SaveConfig(testfix.SignedConfig(a)); err != nil {
		t.Fatal(err)
	}
	if err := d.SaveConfig(testfix.SignedConfig(b)); err != nil {
		t.Fatal(err)
	}
	got, err := d.Config(vps1)
	if err != nil || got.Version != 4 {
		t.Fatalf("Config = %+v, %v", got, err)
	}
	prev, err := os.ReadFile(filepath.Join(d.Path, PrevConfigFile))
	if _, serr := os.Stat(filepath.Join(d.Path, PrevConfigFile+".sig")); err != nil || serr != nil || !strings.Contains(string(prev), `"version":3`) {
		t.Fatalf("config.prev.json = %s, %v, %v", prev, err, serr)
	}
	if err := d.ClearConfig(); err != nil {
		t.Fatal(err)
	}
	if c, err := d.Config(vps1); c != nil || err != nil {
		t.Fatalf("Config after ClearConfig = %v, %v", c, err)
	}
	broken := proto.SignedBody{Body: []byte("{"), Nonce: testfix.Nonce}
	broken.Signature = proto.SignBody(testfix.ConfigKey(), proto.SignedConfig, vps1.NodeID, 0, testfix.Nonce, broken.Body)
	if err := d.SaveConfig(broken); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Config(vps1); err == nil || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("damaged cache = %v", err)
	}
}

// A cache is used only as the control plane signed it for this node
// (amendment A50): changed on disk, without its signature, or another
// node's, it is ignored and says why.
func TestCachesAreCheckedWithThePinnedKey(t *testing.T) {
	d := open(t)
	if err := d.SaveConfig(testfix.SignedConfig(testfix.ServerConfig())); err != nil {
		t.Fatal(err)
	}
	if err := d.SaveMesh(testfix.Signed(proto.SignedMesh, vps1.NodeID, testfix.MeshView()), testfix.Epoch.Unix()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Config(testfix.Pinned(testfix.VPS2.ID)); err == nil || !strings.Contains(err.Error(), "not as the control plane signed it") {
		t.Fatalf("vps-1's config read as vps-2's = %v", err)
	}
	if _, err := d.Mesh(proto.Pinned{Key: vps1.Key, NodeID: vps1.NodeID, Slot: 1}); err == nil {
		t.Fatal("slot 0's mesh view read as slot 1's")
	}
	b, err := os.ReadFile(filepath.Join(d.Path, ConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(b), testfix.VPS2.Endpoint, "198.51.100.66:51820", 1)
	if err := os.WriteFile(filepath.Join(d.Path, ConfigFile), []byte(tampered), 0o600); err != nil {
		t.Fatal(err)
	}
	if c, err := d.Config(vps1); c != nil || err == nil || !strings.Contains(err.Error(), ConfigFile) {
		t.Fatalf("a tampered cache = %+v, %v", c, err)
	}
	if err := os.Remove(filepath.Join(d.Path, MeshFile+".sig")); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Mesh(vps1); err == nil || !strings.Contains(err.Error(), MeshFile) {
		t.Fatalf("a mesh view without its signature = %v", err)
	}
}

// One process per data dir (amendment A50): a second Lock is refused,
// naming the holder's pid; once released, the next one gets it.
func TestLock(t *testing.T) {
	d := open(t)
	unlock, err := d.Lock()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (&Dir{Path: d.Path}).Lock(); !errors.Is(err, ErrLocked) || !strings.Contains(err.Error(), "pid") {
		t.Fatalf("a second Lock = %v", err)
	}
	unlock()
	again, err := d.Lock()
	if err != nil {
		t.Fatalf("Lock after unlock = %v", err)
	}
	again()
	if _, err := (&Dir{Path: filepath.Join(d.Path, "missing")}).Lock(); err == nil || errors.Is(err, ErrLocked) {
		t.Fatalf("Lock in a missing dir = %v", err)
	}
}

// A pending config sits beside the confirmed one and never replaces it.
func TestPendingConfigIsKeptApart(t *testing.T) {
	d := open(t)
	confirmed, pending := testfix.ServerConfig(), testfix.ServerConfig()
	pending.Version = 4
	if err := d.SaveConfig(testfix.SignedConfig(confirmed)); err != nil {
		t.Fatal(err)
	}
	if p, err := d.PendingConfig(vps1); p != nil || err != nil {
		t.Fatalf("PendingConfig before any = %v, %v", p, err)
	}
	if err := d.SavePending(testfix.SignedConfig(pending)); err != nil {
		t.Fatal(err)
	}
	p, err := d.PendingConfig(vps1)
	c, _ := d.Config(vps1)
	if err != nil || p.Version != 4 || c.Version != 3 {
		t.Fatalf("pending %+v, %v; confirmed v%d", p, err, c.Version)
	}
	if err := d.ClearPending(); err != nil {
		t.Fatal(err)
	}
	if err := d.ClearPending(); err != nil {
		t.Fatalf("clearing twice = %v", err)
	}
	if p, _ := d.PendingConfig(vps1); p != nil {
		t.Fatal("the pending config survived ClearPending")
	}
	if err := d.SavePending(testfix.SignedConfig(pending)); err != nil {
		t.Fatal(err)
	}
	if err := d.ClearConfig(); err != nil {
		t.Fatal(err)
	}
	if p, _ := d.PendingConfig(vps1); p != nil {
		t.Fatal("ClearConfig (a re-join) must drop the pending config too")
	}
}

func TestOpenValidatesThePath(t *testing.T) {
	if _, err := Open("data?x=1"); err == nil {
		t.Fatal("Open accepted a path with '?'")
	}
}

// mkdirs puts a non-empty directory where each named file belongs: it can be
// neither read, replaced nor removed.
func mkdirs(t *testing.T, d *Dir, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.MkdirAll(filepath.Join(d.Path, n, "x"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
}

// A file the agent can neither read, replace nor remove is an error that
// names it, and leaves the rest of the data dir as it was.
func TestUnusableFilesAreReported(t *testing.T) {
	d := open(t)
	mkdirs(t, d, EnrollmentFile, ConfigFile, PendingConfigFile, IdentityKeyFile)
	_, errEnrollment := d.Enrollment()
	_, errConfig := d.Config(vps1)
	_, errKeys := d.LoadKeys()
	for name, err := range map[string]error{
		EnrollmentFile:    errEnrollment,
		ConfigFile:        errConfig,
		PendingConfigFile: d.ClearPending(),
		IdentityKeyFile:   errKeys,
	} {
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := d.ClearConfig(); err == nil || !strings.Contains(err.Error(), ConfigFile) {
		t.Errorf("ClearConfig = %v", err)
	}

	// The previous config cannot be kept: the cached one stays as it was.
	d2 := open(t)
	first := testfix.ServerConfig()
	if err := d2.SaveConfig(testfix.SignedConfig(first)); err != nil {
		t.Fatal(err)
	}
	mkdirs(t, d2, PrevConfigFile)
	next := first
	next.Version++
	if err := d2.SaveConfig(testfix.SignedConfig(next)); err == nil {
		t.Fatal("SaveConfig succeeded without keeping the previous config")
	}
	d3 := open(t)
	mkdirs(t, d3, ConfigFile+sigSuffix)
	if err := d3.SaveConfig(testfix.SignedConfig(first)); err == nil {
		t.Fatal("SaveConfig succeeded without its signature")
	}
	if c, _ := d2.Config(vps1); c == nil || c.Version != first.Version {
		t.Fatalf("cached config %+v, want v%d", c, first.Version)
	}

	// A data dir under a regular file.
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(filepath.Join(file, "data")); err == nil || !strings.Contains(err.Error(), "--data") {
		t.Errorf("Open under a file = %v", err)
	}
	under := &Dir{Path: filepath.Join(file, "data")}
	if err := under.SaveKeys(NewKeys()); err == nil {
		t.Error("SaveKeys under a file succeeded")
	}
}

// The node's own settings: none is DNS on, a saved one survives a restart,
// and a damaged or unreadable file says what to do.
func TestLocalSettings(t *testing.T) {
	d := open(t)
	if l, err := d.Local(); err != nil || l.DNSOff {
		t.Fatalf("no file = %+v, %v", l, err)
	}
	if err := d.SaveLocal(Local{DNSOff: true}); err != nil {
		t.Fatal(err)
	}
	if l, err := (&Dir{Path: d.Path}).Local(); err != nil || !l.DNSOff {
		t.Fatalf("after a restart = %+v, %v", l, err)
	}
	// File modes mean nothing on Windows; the data dir's ACL guards it there.
	if st, err := os.Stat(filepath.Join(d.Path, LocalFile)); err != nil || (runtime.GOOS != "windows" && st.Mode().Perm() != 0o600) {
		t.Fatalf("mode = %v, %v", st.Mode().Perm(), err)
	}
	if err := os.WriteFile(filepath.Join(d.Path, LocalFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Local(); err == nil || !strings.Contains(err.Error(), "remove it") {
		t.Fatalf("damaged = %v", err)
	}
	d2 := open(t)
	mkdirs(t, d2, LocalFile)
	if _, err := d2.Local(); err == nil || !strings.Contains(err.Error(), LocalFile) {
		t.Fatalf("unreadable = %v", err)
	}
	if err := d2.SaveLocal(Local{}); err == nil {
		t.Fatal("SaveLocal over a directory succeeded")
	}
}

// The mesh view is cached as fetched, without the reason it was stale, and
// a cache that no longer validates is reported rather than shown.
func TestJoinRecord(t *testing.T) {
	d := open(t)
	if r, err := d.JoinRecord(); err != nil || r.Token != "" {
		t.Fatalf("no file = %+v, %v", r, err)
	}
	want := JoinRecord{Token: strings.Repeat("ab", 32)}
	if err := d.SaveJoinRecord(want); err != nil {
		t.Fatal(err)
	}
	if r, err := (&Dir{Path: d.Path}).JoinRecord(); err != nil || r != want {
		t.Fatalf("after a restart = %+v, %v", r, err)
	}
	if st, err := os.Stat(filepath.Join(d.Path, JoinFile)); err != nil || (runtime.GOOS != "windows" && st.Mode().Perm() != 0o600) {
		t.Fatalf("mode = %v, %v", st.Mode().Perm(), err)
	}
	if err := os.WriteFile(filepath.Join(d.Path, JoinFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := d.JoinRecord(); err == nil || !strings.Contains(err.Error(), "remove it") {
		t.Fatalf("damaged = %v", err)
	}
	d2 := open(t)
	mkdirs(t, d2, JoinFile)
	if _, err := d2.JoinRecord(); err == nil || !strings.Contains(err.Error(), JoinFile) {
		t.Fatalf("unreadable = %v", err)
	}
	if err := d2.SaveJoinRecord(JoinRecord{}); err == nil {
		t.Fatal("SaveJoinRecord over a directory succeeded")
	}
}

func TestMeshCache(t *testing.T) {
	d := open(t)
	if m, err := d.Mesh(vps1); err != nil || m.FetchedAt != 0 || len(m.View.Nodes) != 0 {
		t.Fatalf("no file = %+v, %v", m, err)
	}
	view := testfix.MeshView()
	if err := d.SaveMesh(testfix.Signed(proto.SignedMesh, vps1.NodeID, view), testfix.Epoch.Unix()); err != nil {
		t.Fatal(err)
	}
	m, err := (&Dir{Path: d.Path}).Mesh(vps1)
	if err != nil || m.FetchedAt != testfix.Epoch.Unix() || m.Error != "" || len(m.View.Nodes) != len(view.Nodes) || m.View.Nodes[4] != view.Nodes[4] {
		t.Fatalf("after a restart = %+v, %v", m, err)
	}
	if st, err := os.Stat(filepath.Join(d.Path, MeshFile)); err != nil || (runtime.GOOS != "windows" && st.Mode().Perm() != 0o600) {
		t.Fatalf("mode = %v, %v", st.Mode().Perm(), err)
	}
	for _, bad := range []string{"{", `{"nodes":[{"name":"vps-1;id","role":"server","mesh_ip":"10.77.0.1","health":"ok","state":""}]}`} {
		b := proto.SignedBody{Body: []byte(bad), Nonce: testfix.Nonce}
		b.Signature = proto.SignBody(testfix.ConfigKey(), proto.SignedMesh, vps1.NodeID, 0, testfix.Nonce, b.Body)
		if err := d.SaveMesh(b, 1); err != nil {
			t.Fatal(err)
		}
		if _, err := d.Mesh(vps1); err == nil || !strings.Contains(err.Error(), "is damaged") {
			t.Fatalf("damaged %s = %v", bad, err)
		}
	}
	d2 := open(t)
	mkdirs(t, d2, MeshFile)
	if _, err := d2.Mesh(vps1); err == nil || !strings.Contains(err.Error(), MeshFile) {
		t.Fatalf("unreadable = %v", err)
	}
	if err := d2.SaveMesh(testfix.Signed(proto.SignedMesh, vps1.NodeID, view), 1); err == nil {
		t.Fatal("SaveMesh over a directory succeeded")
	}
}
