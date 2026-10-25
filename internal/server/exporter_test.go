package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestNextExport(t *testing.T) {
	loc := time.FixedZone("UTC+8", 8*3600)
	for in, want := range map[string]string{
		"2026-10-12T02:59:59": "2026-10-12T03:00:00",
		"2026-10-12T03:00:00": "2026-10-13T03:00:00",
		"2026-10-12T23:00:00": "2026-10-13T03:00:00",
		"2026-12-31T04:00:00": "2027-01-01T03:00:00",
	} {
		now, _ := time.ParseInLocation("2006-01-02T15:04:05", in, loc)
		if got := NextExport(now).Format("2006-01-02T15:04:05"); got != want {
			t.Errorf("NextExport(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestExportOnceSkipsUnchangedAndKeepsFourteen(t *testing.T) {
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), DBFile), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	dir := filepath.Join(t.TempDir(), ExportsDir)
	e := &Exporter{DB: st, Dir: dir, Now: clock.Now}

	written, err := e.ExportOnce(context.Background())
	if err != nil || len(written) != 1 || written[0] != filepath.Join(dir, "default", "mesh-2026-10-12.yaml") {
		t.Fatalf("first export = %q, %v", written, err)
	}
	first := written[0]
	if fi, err := os.Stat(first); err != nil || (runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600) {
		t.Fatalf("export file %v, %v; want mode 0600", fi, err)
	}
	clock.Advance(24 * time.Hour)
	if again, err := e.ExportOnce(context.Background()); err != nil || len(again) != 0 {
		t.Fatalf("an unchanged mesh = %q, %v; want no new file", again, err)
	}

	// Sixteen days of changes leave the newest fourteen files.
	for day := 0; day < 16; day++ {
		clock.Advance(24 * time.Hour)
		if err := st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
			return tx.SetSetting(store.SettingDefaultMTU, fmt.Sprint(1300+day))
		}); err != nil {
			t.Fatal(err)
		}
		if f, err := e.ExportOnce(context.Background()); err != nil || len(f) != 1 {
			t.Fatalf("day %d: %q, %v", day, f, err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, "default", "mesh-*.yaml"))
	if len(files) != keepExports {
		t.Fatalf("%d export files, want %d", len(files), keepExports)
	}
	if _, err := os.Stat(first); err == nil {
		t.Fatal("the oldest export was not pruned")
	}
}

func TestExporterRunStopsWithItsContext(t *testing.T) {
	e := &Exporter{Dir: t.TempDir(), Now: time.Now}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		e.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop")
	}
}

// syncBuffer is a log sink Run's goroutine and the test share.
type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// At the export hour Run writes the meshes' exports and logs each file; a
// failed export is logged, and Run carries on until its context ends.
func TestExporterRunExportsAtNight(t *testing.T) {
	// Always 10 ms before the export hour, so the next night is 10 ms away.
	night := time.Date(2026, 10, 12, exportHour-1, 59, 59, 990_000_000, time.UTC)
	now := func() time.Time { return night }
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), DBFile), store.Options{Now: now})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	run := func(db api.DB, want string) string {
		t.Helper()
		var logs syncBuffer
		e := &Exporter{DB: db, Dir: t.TempDir(), Now: now, Log: slog.New(slog.NewTextHandler(&logs, nil))}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			e.Run(ctx)
			close(done)
		}()
		for deadline := time.Now().Add(5 * time.Second); !strings.Contains(logs.String(), want) && time.Now().Before(deadline); {
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
		<-done
		if !strings.Contains(logs.String(), want) {
			t.Errorf("Run's log = %q, want %q", logs.String(), want)
		}
		return e.Dir
	}
	dir := run(st, "nightly export written")
	if files, _ := filepath.Glob(filepath.Join(dir, "default", "mesh-*.yaml")); len(files) != 1 {
		t.Errorf("the night's exports = %q, want one", files)
	}
	run(brokenDB{}, "nightly export failed; it is retried tomorrow")
}

// Each mesh gets its own export directory, named after it (amendment A49).
func TestExportOnceWritesEachMesh(t *testing.T) {
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), DBFile), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Tx(context.Background(), func(tx *store.Tx) error {
		m := store.Mesh{Name: "lab", Slot: 1}
		if err := tx.InsertMesh(&m); err != nil {
			return err
		}
		return tx.In(m.ID).SetSetting(store.SettingMeshCIDR, "10.78.0.0/16")
	}); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), ExportsDir)
	written, err := (&Exporter{DB: st, Dir: dir, Now: clock.Now}).ExportOnce(context.Background())
	want := []string{filepath.Join(dir, "default", "mesh-2026-10-12.yaml"), filepath.Join(dir, "lab", "mesh-2026-10-12.yaml")}
	if err != nil || len(written) != 2 || written[0] != want[0] || written[1] != want[1] {
		t.Fatalf("ExportOnce = %q, %v; want %q", written, err, want)
	}
	lab, err := os.ReadFile(want[1])
	if err != nil || !strings.Contains(string(lab), "10.78.0.0/16") || strings.Contains(string(lab), "10.77.0.0/16") {
		t.Fatalf("the lab's export: %v\n%s", err, lab)
	}
}

// brokenDB fails every transaction, as a broken disk would.
type brokenDB struct{}

func (brokenDB) Tx(context.Context, func(*store.Tx) error) error {
	return errors.New("disk I/O error")
}

// A failing store, or a mesh directory that cannot be made, is an error.
func TestExportOnceFails(t *testing.T) {
	if _, err := (&Exporter{DB: brokenDB{}, Dir: t.TempDir(), Now: time.Now}).ExportOnce(context.Background()); err == nil {
		t.Fatal("a failing store exported")
	}
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), DBFile), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	file := filepath.Join(t.TempDir(), "exports")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	written, err := (&Exporter{DB: st, Dir: file, Now: clock.Now}).ExportOnce(context.Background())
	if err == nil || len(written) != 0 || !strings.Contains(err.Error(), "export mesh default") {
		t.Fatalf("an export into a file = %q, %v", written, err)
	}
}
