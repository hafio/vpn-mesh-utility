// Package wintun ships wintun.dll, the driver wireguard-go needs on Windows
// (amendment A34). The release build embeds it (build tag wintun, after
// internal/tools/fetchwintun fetched and checked it); at start the agent
// writes it beside meshctl.exe, where Windows looks for it, and checks it
// there against the same pin every time.
//
// The zip and each DLL are pinned by SHA-256: a DLL that does not match
// is never written or loaded.
package wintun

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
)

// The release this build ships, as wintun.net publishes it.
const (
	Version   = "0.14.1"
	ZipURL    = "https://www.wintun.net/builds/wintun-0.14.1.zip"
	ZipSHA256 = "07c256185d6ee3652e09fa55c0b673e2624b565e02c4b9091c79ca7d2f24ef51"
	// MaxZip and MaxDLL bound what is read from the download.
	MaxZip = 4 << 20
	MaxDLL = 2 << 20
)

// Pin is one DLL in the zip.
type Pin struct {
	Path   string
	SHA256 string
}

// DLLs are the drivers in the zip, by GOARCH.
var DLLs = map[string]Pin{
	"amd64": {Path: "wintun/bin/amd64/wintun.dll", SHA256: "e5da8447dc2c320edc0fc52fa01885c103de8c118481f683643cacc3220dafce"},
	"arm64": {Path: "wintun/bin/arm64/wintun.dll", SHA256: "f7ba89005544be9d85231a9e0d5f23b2d15b3311667e2dad0debd344918a3f80"},
}

// FileName is the DLL's name beside meshctl.exe, as wireguard-go loads it.
const FileName = "wintun.dll"

// ErrNotBundled means this binary was built without the DLL.
var ErrNotBundled = errors.New("this meshctl was built without wintun.dll, so it cannot run WireGuard on Windows: " +
	"use a release binary, or build with ./scripts/dev.ps1 build (it fetches and checks the DLL)")

// Sum is b's SHA-256 in hex.
func Sum(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Ensure writes the bundled DLL beside the running binary unless the one
// there already matches the pin, and says where it is and whether it wrote
// it.
func Ensure() (string, bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", false, fmt.Errorf("find meshctl.exe for wintun.dll: %w", err)
	}
	pin, ok := DLLs[runtime.GOARCH]
	if !ok {
		return "", false, fmt.Errorf("wintun.dll: no driver for %s", runtime.GOARCH)
	}
	return install(filepath.Dir(exe), bundled(), pin.SHA256)
}

// install puts dll in dir as FileName, checked against want.
func install(dir string, dll []byte, want string) (string, bool, error) {
	path := filepath.Join(dir, FileName)
	if len(dll) == 0 {
		return path, false, ErrNotBundled
	}
	if got := Sum(dll); got != want {
		return path, false, fmt.Errorf("the wintun.dll in this build is not the pinned %s (sha256 %s); rebuild it", Version, got)
	}
	if cur, err := os.ReadFile(path); err == nil && Sum(cur) == want {
		return path, false, nil
	}
	if err := fsutil.WriteAtomic(dir, FileName, dll, 0o644); err != nil {
		return path, false, fmt.Errorf("write %s: %w (is another program using it?)", path, err)
	}
	return path, true, nil
}
