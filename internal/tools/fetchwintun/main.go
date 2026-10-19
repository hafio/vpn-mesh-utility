// Command fetchwintun puts wintun.dll beside internal/agent/wintun for the
// Windows release build (amendment A34): it downloads the pinned wintun
// release, checks the zip and the DLL against their SHA-256 pins, and
// writes wintun_<arch>.dll. A DLL already there that matches is kept, so a
// rebuild needs no network.
//
// Usage (the build and dist tasks of scripts/dev.sh and scripts/dev.ps1,
// and the Dockerfile, run it before building windows/amd64 with -tags
// wintun):
//
//	go run ./internal/tools/fetchwintun -arch amd64 -out internal/agent/wintun
//
// Exit codes: 0 the DLL is in place, 1 it could not be fetched or did not
// match its pin, 2 bad flags.
package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/wintun"
	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, download)) }

// run is the command with its flags, output and download as seams.
func run(args []string, stdout, stderr io.Writer, get func(string) ([]byte, error)) int {
	fs := flag.NewFlagSet("fetchwintun", flag.ContinueOnError)
	fs.SetOutput(stderr)
	arch := fs.String("arch", "amd64", "the Windows architecture to fetch the driver for (amd64 or arm64)")
	out := fs.String("out", filepath.Join("internal", "agent", "wintun"), "where to write wintun_<arch>.dll")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	pin, ok := wintun.DLLs[*arch]
	if !ok || fs.NArg() != 0 {
		fmt.Fprintln(stderr, "fetchwintun: -arch must be amd64 or arm64, with no arguments")
		return 2
	}
	path, fetched, err := fetch(get, wintun.ZipSHA256, pin, *out, "wintun_"+*arch+".dll")
	switch {
	case err != nil:
		fmt.Fprintf(stderr, "fetchwintun: %v\n", err)
		return 1
	case fetched:
		fmt.Fprintf(stdout, "fetched wintun %s into %s\n", wintun.Version, path)
	default:
		fmt.Fprintf(stdout, "%s is already the pinned wintun %s\n", path, wintun.Version)
	}
	return 0
}

// fetch writes the DLL pin names to dir/name, from the zip get returns
// (checked against zipSum), unless the file there already matches. It
// reports whether it fetched.
func fetch(get func(url string) ([]byte, error), zipSum string, pin wintun.Pin, dir, name string) (string, bool, error) {
	path := filepath.Join(dir, name)
	if cur, err := os.ReadFile(path); err == nil && wintun.Sum(cur) == pin.SHA256 {
		return path, false, nil
	}
	z, err := get(wintun.ZipURL)
	if err != nil {
		return path, false, err
	}
	dll, err := extract(z, zipSum, pin)
	if err != nil {
		return path, false, err
	}
	if err := fsutil.WriteAtomic(dir, name, dll, 0o644); err != nil {
		return path, false, err
	}
	return path, true, nil
}

// extract checks the zip against zipSum, then returns the DLL pin names
// once it matches its own.
func extract(z []byte, zipSum string, pin wintun.Pin) ([]byte, error) {
	if got := wintun.Sum(z); got != zipSum {
		return nil, fmt.Errorf("%s has sha256 %s, not the pinned %s: not using it", wintun.ZipURL, got, zipSum)
	}
	r, err := zip.NewReader(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", wintun.ZipURL, err)
	}
	f, err := r.Open(pin.Path)
	if err != nil {
		return nil, fmt.Errorf("%s has no %s: %w", wintun.ZipURL, pin.Path, err)
	}
	defer f.Close()
	dll, err := io.ReadAll(io.LimitReader(f, wintun.MaxDLL+1))
	switch {
	case err != nil:
		return nil, fmt.Errorf("read %s: %w", pin.Path, err)
	case len(dll) > wintun.MaxDLL:
		return nil, fmt.Errorf("%s is over %d bytes: not using it", pin.Path, wintun.MaxDLL)
	case wintun.Sum(dll) != pin.SHA256:
		return nil, fmt.Errorf("%s has sha256 %s, not the pinned %s: not using it", pin.Path, wintun.Sum(dll), pin.SHA256)
	}
	return dll, nil
}

// download gets url over https, at most wintun.MaxZip bytes.
func download(url string) ([]byte, error) {
	c := &http.Client{Timeout: time.Minute}
	resp, err := c.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, wintun.MaxZip+1))
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	if len(b) > wintun.MaxZip {
		return nil, errors.New("download " + url + ": over the size limit")
	}
	return b, nil
}
