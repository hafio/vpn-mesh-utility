package api

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Downloads (amendment A14): the image bakes the binaries for every shipped
// target and their SHA256SUMS into MESHCTL_DIST_DIR; the server serves them
// by exact name only and renders install.sh with the checksums pinned in it.

var (
	reDistFile = regexp.MustCompile(`^(meshctl-linux-amd64|meshctl-linux-arm64|meshctl-windows-amd64\.exe|SHA256SUMS)$`)
	reSumLine  = regexp.MustCompile(`^([0-9a-f]{64})  (\S+)$`)
)

const sumsFile = "SHA256SUMS"

// downloadTimeout replaces the default per-response write deadline: a 15 MB
// binary over a slow link takes minutes.
const downloadTimeout = 10 * time.Minute

func (a *api) noDownloads() error {
	return errorf(http.StatusNotFound, proto.CodeNotFound, "this server has no downloads: MESHCTL_DIST_DIR is not set (the release image sets it)")
}

func (a *api) download(w http.ResponseWriter, r *http.Request) {
	if a.DistDir == "" {
		a.fail(w, r, a.noDownloads())
		return
	}
	name := r.PathValue("file")
	if !reDistFile.MatchString(name) {
		a.fail(w, r, errorf(http.StatusNotFound, proto.CodeNotFound,
			"no download named %q; available: meshctl-linux-amd64, meshctl-linux-arm64, meshctl-windows-amd64.exe, SHA256SUMS", name))
		return
	}
	f, err := os.Open(filepath.Join(a.DistDir, name))
	if err != nil {
		a.fail(w, r, errorf(http.StatusNotFound, proto.CodeNotFound, "%s is not in this server's dist directory", name))
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		a.fail(w, r, fmt.Errorf("stat %s: %w", name, err))
		return
	}
	ct := "application/octet-stream"
	if name == sumsFile {
		ct = "text/plain; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(downloadTimeout)) // unsupported writers keep the server's timeouts
	http.ServeContent(w, r, name, st.ModTime(), f)
}

func (a *api) install(w http.ResponseWriter, r *http.Request) {
	a.script(w, r, deploy.RenderInstall, "text/x-shellscript; charset=utf-8")
}

// installPS1 is the Windows installer (amendment A51).
func (a *api) installPS1(w http.ResponseWriter, r *http.Request) {
	a.script(w, r, deploy.RenderInstallPS1, "text/plain; charset=utf-8")
}

// script answers an install script rendered with this server's origin,
// version and the checksums of its binaries.
func (a *api) script(w http.ResponseWriter, r *http.Request, render func(deploy.InstallParams) ([]byte, error), contentType string) {
	if a.DistDir == "" {
		a.fail(w, r, a.noDownloads())
		return
	}
	sums, err := readSums(filepath.Join(a.DistDir, sumsFile))
	if err != nil {
		a.fail(w, r, err)
		return
	}
	script, err := render(deploy.InstallParams{
		Origin: a.ServerURL, Version: a.Version,
		SHA256Amd64: sums["meshctl-linux-amd64"], SHA256Arm64: sums["meshctl-linux-arm64"], SHA256Windows: sums["meshctl-windows-amd64.exe"],
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(script) // the client went away; nothing to do
}

// readSums parses sha256sum output; lines that are not "<hex>  <name>" are
// ignored, and a missing entry surfaces when install.sh is rendered.
func readSums(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read checksums: %w (the image writes SHA256SUMS next to the binaries)", err)
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if m := reSumLine.FindStringSubmatch(sc.Text()); m != nil {
			out[m[2]] = m[1]
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read checksums %s: %w", path, err)
	}
	return out, nil
}
