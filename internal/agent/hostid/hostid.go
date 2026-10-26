// Package hostid reads the host's machine id (amendment A50), which pins a
// node to the machine it joined on: a copy of its data dir on another
// machine is refused.
//
// The id itself never leaves the host. The agent tells the control plane
// Presented, a hash of it that names meshctl (as systemd advises for
// machine ids), and keeps Local, another hash, in the data dir, so a copy
// of the data dir learns nothing it could present.
package hostid

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

// maxID bounds what is read as a machine id.
const maxID = 256

// Read returns the host's machine id: the file at path when one is given,
// else the platform's (sources).
func Read(path string) (string, error) {
	if path != "" {
		return fromFile(path)
	}
	return first(sources())
}

// first is the id of the first source that holds one.
func first(srcs []source) (string, error) {
	var tried []string
	for _, s := range srcs {
		id, err := s.read()
		if err == nil {
			return id, nil
		}
		tried = append(tried, s.name)
	}
	return "", fmt.Errorf("this host has no machine id (looked in %s); in a container, mount the host's "+
		"(`/etc/machine-id:/etc/machine-id:ro`), or name a file with --machine-id-file", strings.Join(tried, ", "))
}

// source is one place a platform keeps its machine id.
type source struct {
	name string
	read func() (string, error)
}

func fromFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read the machine id: %w", err)
	}
	id, err := clean(string(b))
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, err)
	}
	return id, nil
}

// errNoID is a source that holds no usable id.
var errNoID = errors.New("holds no machine id")

// clean is s as a machine id: trimmed, printable, at most maxID bytes, and
// not a placeholder (empty, all zeros, systemd's "uninitialized").
func clean(s string) (string, error) {
	s = strings.TrimSpace(s)
	placeholder := strings.Trim(s, "0-") == "" || s == "uninitialized"
	if placeholder || len(s) > maxID || strings.ContainsFunc(s, func(r rune) bool { return r < 0x21 || r > 0x7e }) {
		return "", errNoID
	}
	return s, nil
}

// Presented is what the agent tells the control plane of id.
func Presented(id string) string { return hash("meshctl machine id\x00", id) }

// Local is what the data dir keeps of id.
func Local(id string) string { return hash("meshctl data dir machine\x00", id) }

func hash(prefix, id string) string {
	sum := sha256.Sum256([]byte(prefix + id))
	return hex.EncodeToString(sum[:])
}
