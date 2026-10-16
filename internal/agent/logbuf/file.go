package logbuf

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// File is a log file that stays under max bytes: when a write would pass
// it, the file moves to <path>.1 (replacing the one before) and starts
// again. The Windows service logs here, since nothing reads its stderr.
type File struct {
	mu   sync.Mutex
	path string
	max  int64
	f    *os.File
	size int64
}

// OpenFile appends to the log at path.
func OpenFile(path string, max int64) (*File, error) {
	l := &File{path: path, max: max}
	if err := l.open(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *File) open() error {
	f, err := os.OpenFile(l.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %s: %w", l.path, err)
	}
	st, err := f.Stat()
	if err != nil {
		_ = f.Close() // failing anyway
		return fmt.Errorf("open log file %s: %w", l.path, err)
	}
	l.f, l.size = f, st.Size()
	return nil
}

// Write appends p, rotating first when p would take the file past max. A
// rotation that fails is reported, but the line is still written to the
// file as it is, so the log goes on.
func (l *File) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	var rerr error
	if l.size > 0 && l.size+int64(len(p)) > l.max {
		rerr = l.rotate()
	}
	if l.f == nil {
		return 0, rerr
	}
	n, err := l.f.Write(p)
	l.size += int64(n)
	if err == nil {
		err = rerr
	}
	return n, err
}

// rotate moves the file to <path>.1 and opens a new one; when the move
// fails, it opens the old one again.
func (l *File) rotate() error {
	err := l.f.Close()
	l.f = nil
	old := l.path + ".1"
	if err == nil {
		if err = os.Remove(old); errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
	}
	if err == nil {
		err = os.Rename(l.path, old)
	}
	if oerr := l.open(); oerr != nil {
		return oerr
	}
	if err != nil {
		return fmt.Errorf("rotate log file %s: %w", l.path, err)
	}
	return nil
}

// Close closes the file.
func (l *File) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f == nil {
		return nil
	}
	err := l.f.Close()
	l.f = nil
	return err
}
