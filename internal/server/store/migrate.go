package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

// loadMigrations reads the embedded NNNN_name.sql files, in order. Versions
// must start at 1 and be contiguous, so a missing or duplicated file fails
// loudly at startup instead of silently skipping a schema change.
func loadMigrations(fsys fs.FS) ([]migration, error) {
	files, err := fs.Glob(fsys, "migrations/*.sql")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}
	var ms []migration
	for _, f := range files {
		base := strings.TrimSuffix(strings.TrimPrefix(f, "migrations/"), ".sql")
		num, name, ok := strings.Cut(base, "_")
		v, err := strconv.Atoi(num)
		if !ok || err != nil || v <= 0 {
			return nil, fmt.Errorf("migration %s: file name must be NNNN_name.sql", f)
		}
		b, err := fs.ReadFile(fsys, f)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", f, err)
		}
		ms = append(ms, migration{version: v, name: name, sql: string(b)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].version < ms[j].version })
	for i, m := range ms {
		if m.version != i+1 {
			return nil, fmt.Errorf("migrations must be numbered 1..N without gaps; expected %04d, found %04d_%s", i+1, m.version, m.name)
		}
	}
	return ms, nil
}

// migrate brings the schema up to the newest embedded migration. Each
// migration and its schema_version row commit together, so a crash leaves
// the database at a clean version.
func (s *Store) migrate(ctx context.Context) error {
	ms, err := loadMigrations(migrationFS)
	if err != nil {
		return err
	}
	return s.migrateTo(ctx, ms)
}

func (s *Store) migrateTo(ctx context.Context, ms []migration) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_version (
		version    INTEGER NOT NULL PRIMARY KEY,
		applied_at INTEGER NOT NULL
	) STRICT`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}
	var current int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	latest := 0
	if len(ms) > 0 {
		latest = ms[len(ms)-1].version
	}
	if current > latest {
		return fmt.Errorf("database schema is v%d but this meshctl only knows up to v%d: upgrade meshctl, or restore the backup taken before the upgrade", current, latest)
	}
	for _, m := range ms {
		if m.version <= current {
			continue
		}
		err := s.Tx(ctx, func(t *Tx) error {
			if _, err := t.exec(m.sql); err != nil {
				return fmt.Errorf("apply migration %04d_%s: %w", m.version, m.name, err)
			}
			_, err := t.exec(`INSERT INTO schema_version (version, applied_at) VALUES (?, ?)`, m.version, t.now.Unix())
			return err
		})
		if err != nil {
			return err
		}
		s.log.Info("applied database migration", "version", m.version, "name", m.name)
	}
	return nil
}

// checkSchema is the read-only counterpart of migrate: the database must be
// at exactly this binary's schema, because a reader cannot migrate it.
func (s *Store) checkSchema(ctx context.Context) error {
	ms, err := loadMigrations(migrationFS)
	if err != nil {
		return err
	}
	latest := ms[len(ms)-1].version
	var current int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&current); err != nil {
		return fmt.Errorf("read schema version: %w (is this a meshctl database?)", err)
	}
	if current != latest {
		return fmt.Errorf("database schema is v%d but this meshctl expects v%d: run the command with the server's own meshctl binary", current, latest)
	}
	return nil
}

// SchemaVersion reports the applied schema version.
func (t *Tx) SchemaVersion() (int, error) {
	var v int
	err := t.queryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&v)
	return v, err
}
