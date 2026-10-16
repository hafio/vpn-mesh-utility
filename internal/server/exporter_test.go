package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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

	first, err := e.ExportOnce(context.Background())
	if err != nil || filepath.Base(first) != "mesh-2026-10-12.yaml" {
		t.Fatalf("first export = %q, %v", first, err)
	}
	if fi, err := os.Stat(first); err != nil || (runtime.GOOS != "windows" && fi.Mode().Perm() != 0o600) {
		t.Fatalf("export file %v, %v; want mode 0600", fi, err)
	}
	clock.Advance(24 * time.Hour)
	if again, err := e.ExportOnce(context.Background()); err != nil || again != "" {
		t.Fatalf("an unchanged mesh = %q, %v; want no new file", again, err)
	}

	// Sixteen days of changes leave the newest fourteen files.
	for day := 0; day < 16; day++ {
		clock.Advance(24 * time.Hour)
		if err := st.Tx(context.Background(), func(tx *store.Tx) error {
			return tx.SetSetting(store.SettingDefaultMTU, fmt.Sprint(1300+day))
		}); err != nil {
			t.Fatal(err)
		}
		if f, err := e.ExportOnce(context.Background()); err != nil || f == "" {
			t.Fatalf("day %d: %q, %v", day, f, err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(dir, "mesh-*.yaml"))
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
