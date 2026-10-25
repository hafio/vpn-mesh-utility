package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// `agent run --runtime-socket` takes auto, off or a clean absolute path,
// and --status-every 1s to 5m; anything else is refused before the agent
// starts.
func TestAgentRunRuntimeSocket(t *testing.T) {
	data := filepath.Join(shortDir(t), "data")
	for _, bad := range []string{"docker.sock", "/run/../docker.sock", "/run/docker.sock;id", ""} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", "off", "--runtime-socket", bad); code != exitUsage ||
			!strings.Contains(errOut, "--runtime-socket") {
			t.Errorf("--runtime-socket %q = %d %q", bad, code, errOut)
		}
	}
	for _, bad := range []string{"500ms", "6m", "often"} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", "off", "--status-every", bad); code != exitUsage ||
			!strings.Contains(errOut, "status-every") {
			t.Errorf("--status-every %s = %d %q", bad, code, errOut)
		}
	}
	for _, bad := range []string{"4s", "6m", "soon"} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", "off", "--endpoint-retry", bad); code != exitUsage ||
			!strings.Contains(errOut, "endpoint-retry") {
			t.Errorf("--endpoint-retry %s = %d %q", bad, code, errOut)
		}
	}
	for _, good := range []string{"auto", "off", "/run/user/1000/podman/podman.sock"} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", "off", "--runtime-socket", good); code != exitFailure ||
			!strings.Contains(errOut, "meshctl agent join") {
			t.Errorf("--runtime-socket %s before a join = %d %q", good, code, errOut)
		}
	}
}
