// Package server runs the control plane, `meshctl control`: it locks the
// data dir, opens the store, brings every node's computed config in line
// with this binary, and serves the agent API on --listen and the admin API
// on <data>/admin.sock until the context ends, writing a nightly export on
// the side. Import (the offline restore) lives here too, because it shares
// the lock.
package server

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/server/ui"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Files in the data directory.
const (
	DBFile      = "mesh.db"
	AdminSocket = "admin.sock"
)

const (
	shutdownGrace = 10 * time.Second
	// maxSocketPath leaves room under the smallest sun_path (104 bytes).
	maxSocketPath = 100
)

// Config is everything Run needs; cmd/meshctl fills it from flags.
type Config struct {
	DataDir string
	Listen  string
	// URL is the public origin agents and install.sh use.
	URL            string
	DistDir        string
	TrustedProxies []netip.Prefix
	Version        string
	Log            *slog.Logger
	Now            func() time.Time
	// HubDownAfter is how long a hub may be silent before relay failover
	// moves its pairs (amendment A40); zero turns failover off.
	HubDownAfter time.Duration
	// Ready, when set, is called once both listeners accept connections.
	Ready func(addr net.Addr)
}

// Run serves until ctx is cancelled or a listener fails.
func Run(ctx context.Context, c Config) error {
	if c.Log == nil {
		c.Log = slog.New(slog.DiscardHandler)
	}
	if c.Now == nil {
		c.Now = time.Now
	}
	if _, err := validate.FilePath("--data", c.DataDir); err != nil {
		return err
	}
	origin, err := validate.ServerURL("--url", c.URL)
	if err != nil {
		return err
	}
	if c.DistDir != "" {
		if _, err := validate.FilePath("MESHCTL_DIST_DIR", c.DistDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(c.DataDir, 0o700); err != nil {
		return fmt.Errorf("create data dir %s: %w", c.DataDir, err)
	}
	unlock, err := Lock(c.DataDir)
	if err != nil {
		return err
	}
	defer unlock()
	// The key that seals TOTP secrets is created on first start.
	key, err := auth.LoadOrCreateKey(filepath.Join(c.DataDir, auth.KeyFile))
	if err != nil {
		return err
	}
	sealer, err := auth.NewSealer(key)
	if err != nil {
		return err
	}
	// So is the key every config, mesh view and definitive refusal is
	// signed with (amendment A50); agents pin its public half at join.
	seed, err := auth.LoadOrCreateKey(filepath.Join(c.DataDir, api.ConfigKeyFile))
	if err != nil {
		return err
	}
	st, err := store.Open(ctx, filepath.Join(c.DataDir, DBFile), store.Options{Now: c.Now, Log: c.Log})
	if err != nil {
		return err
	}
	defer st.Close()

	hub := api.NewHub()
	rec := api.NewReconciler(st, hub, c.Log)
	// A new release may compute configs differently; reconcile every mesh
	// once so the stored hashes, and the agents, follow this binary -- and
	// the server's time zone, which the mesh firewall's schedules default
	// to.
	zone := serverZone(c.Now().Location(), os.Readlink)
	var meshes []store.Mesh
	if err := st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		if meshes, err = tx.Meshes(); err != nil {
			return err
		}
		return tx.SetServerSetting(store.SettingServerTZ, zone)
	}); err != nil {
		return fmt.Errorf("reconcile at startup: %w", err)
	}
	for _, m := range meshes {
		res, err := rec.Apply(ctx, api.Change{Mesh: m.ID, Actor: "system", Action: "server.start", Summary: "recomputed by meshctl " + c.Version,
			Detail: map[string]any{"version": c.Version}}, nil)
		var pe *api.PlanError
		switch {
		case errors.As(err, &pe):
			c.Log.Error("the stored mesh plan has errors; agents keep their last config until it is fixed with `meshctl node set`",
				"mesh", m.Name, "err", err)
		case err != nil:
			return fmt.Errorf("reconcile mesh %s at startup: %w", m.Name, err)
		}
		for _, w := range res.Warnings {
			c.Log.Warn("mesh plan", "mesh", m.Name, "problem", w.String())
		}
	}

	deps := api.Deps{
		DB: st, Hub: hub, Reconciler: rec, Now: c.Now, Log: c.Log, ServerURL: origin, Version: c.Version,
		DistDir: c.DistDir, TrustedProxies: c.TrustedProxies, ConfigKey: ed25519.NewKeyFromSeed(seed),
	}
	nodes := api.NewNodes(deps)
	// Both families by default (M8): a mesh whose IPv6 was never chosen gets
	// a random unique local prefix; one given up stays IPv4 only.
	for _, m := range meshes {
		switch v, set, err := nodes.In(m.ID).DefaultIPv6(ctx); {
		case err != nil:
			c.Log.Error("the mesh got no IPv6 prefix; it runs IPv4 only until `meshctl ipv6 auto`", "mesh", m.Name, "err", err)
		case set:
			c.Log.Info("the mesh's IPv6 prefix", "mesh", m.Name, "prefix", v.Prefix, "nodes", v.Addressed)
		}
	}
	web, err := ui.New(ui.Deps{
		DB: st, Nodes: nodes, ServerURL: origin, Version: c.Version,
		TrustedProxies: c.TrustedProxies, Now: c.Now, Log: c.Log,
		// The issuer names the mesh in authenticator apps; no ':' (it
		// separates issuer and account in the otpauth label).
		Auth: auth.New(auth.Config{DB: st, Sealer: sealer, Now: c.Now, Issuer: "meshctl " + hostOf(origin)}),
	})
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", c.Listen)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", c.Listen, err)
	}
	adminPath := filepath.Join(c.DataDir, AdminSocket)
	adminLn, err := ListenAdmin(adminPath)
	if err != nil {
		_ = ln.Close() // the admin error is the one worth reporting
		return err
	}
	errLog := slog.NewLogLogger(c.Log.Handler(), slog.LevelWarn)
	// No ReadTimeout or WriteTimeout: both would cut SSE streams. The
	// handlers bound each body read and each write instead.
	pub := &http.Server{Handler: Routes(api.New(deps), web), ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 120 * time.Second,
		ErrorLog: errLog}
	adm := &http.Server{Handler: api.NewAdmin(deps), ReadHeaderTimeout: 10 * time.Second, ErrorLog: errLog}
	errc := make(chan error, 2)
	go func() { errc <- serve(pub, ln) }()
	go func() { errc <- serve(adm, adminLn) }()
	exports := &Exporter{DB: st, Dir: filepath.Join(c.DataDir, ExportsDir), Now: c.Now, Log: c.Log}
	exportCtx, stopExports := context.WithCancel(ctx)
	defer stopExports()
	go exports.Run(exportCtx)
	failover := &api.Failover{DB: st, Reconciler: rec, Now: c.Now, Log: c.Log, DownAfter: c.HubDownAfter}
	go failover.Run(exportCtx)
	c.Log.Info("meshctl control plane listening", "addr", ln.Addr().String(), "url", origin, "admin_socket", adminPath, "version", c.Version)
	if c.Ready != nil {
		c.Ready(ln.Addr())
	}

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errc:
	}
	hub.Close() // end SSE streams, or Shutdown would wait out its grace period
	sctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	return errors.Join(runErr, pub.Shutdown(sctx), adm.Shutdown(sctx))
}

// Routes sends the agent API, health check and downloads to agent and every
// other path to the web UI. A GeoIP allowlist in front of the server may
// restrict the UI; the agent paths must stay reachable from every node.
func Routes(agent, web http.Handler) http.Handler {
	mux := http.NewServeMux()
	for _, p := range []string{proto.AgentPrefix, proto.PathHealth, proto.PathInstall, proto.PathDownload} {
		mux.Handle(p, agent)
	}
	mux.Handle("/", web)
	return mux
}

func hostOf(origin string) string {
	u, err := url.Parse(origin)
	if err != nil {
		return origin
	}
	return u.Hostname()
}

func serve(s *http.Server, ln net.Listener) error {
	if err := s.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// ListenAdmin listens on the admin socket with mode 0600, so only the
// server's own user reaches the admin API. It refuses to start when another
// server already answers there, and removes a stale socket otherwise.
func ListenAdmin(path string) (net.Listener, error) {
	if len(path) > maxSocketPath {
		return nil, fmt.Errorf("admin socket path %s is %d bytes, over the %d a Unix socket allows; use a shorter --data", path, len(path), maxSocketPath)
	}
	if conn, err := net.DialTimeout("unix", path, time.Second); err == nil {
		_ = conn.Close()
		return nil, fmt.Errorf("another meshctl control plane is already running with this data dir (it answers on %s)", path)
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("remove stale admin socket %s: %w", path, err)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("listen on admin socket %s: %w", path, err)
	}
	if runtime.GOOS == "windows" {
		return ln, nil // the server runs on Linux; Windows only builds it
	}
	if err := os.Chmod(path, 0o600); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("restrict admin socket %s to its owner: %w", path, err)
	}
	return ln, nil
}
