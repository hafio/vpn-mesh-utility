// Package store is the server's SQLite persistence (spec S8.1, S8.2): the
// schema and its migrations, and typed accessors used inside one transaction
// at a time.
//
// All access goes through Store.Tx. The pool holds a single connection and
// every transaction begins IMMEDIATE, so writes in this process serialise and
// a concurrent CLI process (`meshctl user add`, `export`) waits on
// busy_timeout instead of failing with SQLITE_BUSY mid-transaction.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

var (
	// ErrNotFound means the requested row does not exist.
	ErrNotFound = errors.New("not found")
	// ErrConflict means a write would violate a uniqueness constraint.
	ErrConflict = errors.New("conflicts with an existing record")
)

// Options configures Open.
type Options struct {
	// Now is the clock every transaction reads once at its start.
	Now func() time.Time
	// Log receives migration progress.
	Log *slog.Logger
	// ReadOnly opens an existing database without migrating it or taking the
	// write lock: `meshctl export` reads a database a running server owns.
	ReadOnly bool
}

// Store is an open database.
type Store struct {
	db  *sql.DB
	now func() time.Time
	log *slog.Logger
}

// dsnParams are the connection settings: WAL for concurrent readers, a busy
// timeout for the CLI-while-server-runs case, foreign keys on (SQLite
// defaults them off), and IMMEDIATE transactions.
const dsnParams = "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)&_txlock=immediate"

// dsnReadOnly refuses every write at the connection and uses deferred
// transactions, so a reader never blocks the server.
const dsnReadOnly = "?_pragma=busy_timeout(5000)&_pragma=query_only(1)"

// Open opens (creating if needed) the database at path and applies pending
// migrations. path is validated because it is embedded in the driver DSN.
func Open(ctx context.Context, path string, opts Options) (*Store, error) {
	if _, err := validate.FilePath("database path", path); err != nil {
		return nil, err
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Log == nil {
		opts.Log = slog.New(slog.DiscardHandler)
	}
	params := dsnParams
	if opts.ReadOnly {
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("open database %s: %w (is --data the server's data directory?)", path, err)
		}
		params = dsnReadOnly
	}
	db, err := sql.Open("sqlite", path+params)
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close() // the ping error is the one worth reporting
		return nil, fmt.Errorf("open database %s: %w (check that the directory exists and is writable)", path, err)
	}
	s := &Store{db: db, now: opts.Now, log: opts.Log}
	check := s.migrate
	if opts.ReadOnly {
		check = s.checkSchema
	}
	if err := check(ctx); err != nil {
		_ = db.Close() // the migration error is the one worth reporting
		return nil, err
	}
	return s, nil
}

// Close closes the database.
func (s *Store) Close() error { return s.db.Close() }

// Tx is one transaction. Its methods are the only way to read or write
// tables; the clock is read once when the transaction starts so every row it
// writes carries the same timestamp.
type Tx struct {
	tx  *sql.Tx
	ctx context.Context
	now time.Time
}

// Now is the transaction's timestamp.
func (t *Tx) Now() time.Time { return t.now }

// Tx runs fn in a transaction, committing if it returns nil and rolling back
// otherwise.
func (s *Store) Tx(ctx context.Context, fn func(*Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin transaction: %w", err)
	}
	t := &Tx{tx: tx, ctx: ctx, now: s.now()}
	if err := fn(t); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("store: rollback: %w", rbErr))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: commit: %w", err)
	}
	return nil
}

func (t *Tx) exec(query string, args ...any) (sql.Result, error) {
	res, err := t.tx.ExecContext(t.ctx, query, args...)
	return res, classify(err)
}

func (t *Tx) query(query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(t.ctx, query, args...)
}

func (t *Tx) queryRow(query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(t.ctx, query, args...)
}

// classify maps uniqueness violations to ErrConflict, keeping the driver
// error in the chain for the log. The bare SQLITE_CONSTRAINT code is matched
// too, in case the driver reports primary rather than extended result codes.
func classify(err error) error {
	if err == nil {
		return nil
	}
	var se *sqlite.Error
	if errors.As(err, &se) {
		switch se.Code() {
		case sqlite3.SQLITE_CONSTRAINT_UNIQUE, sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY, sqlite3.SQLITE_CONSTRAINT:
			return fmt.Errorf("%w: %w", ErrConflict, err)
		}
	}
	return err
}

// nullInt maps 0 to SQL NULL for optional foreign keys and timestamps.
func nullInt(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

// nullStr maps "" to SQL NULL for optional unique columns.
func nullStr(v string) sql.NullString {
	return sql.NullString{String: v, Valid: v != ""}
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
