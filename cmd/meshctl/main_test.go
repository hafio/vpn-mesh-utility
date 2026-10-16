package main

import (
	"bytes"
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// testEnv is an env with captured output, no environment variables and the
// fixed test clock.
func testEnv(getenv map[string]string) (*env, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	return &env{version: "v9.8.7", stdin: strings.NewReader(""), stdout: &out, stderr: &errb, goos: runtime.GOOS,
		getenv: func(k string) string { return getenv[k] }, now: func() time.Time { return testfix.Epoch },
		password: func(string) (string, error) { return "", errors.New("no terminal in tests") }}, &out, &errb
}

func runArgs(args ...string) (code int, stdout, stderr string) {
	e, out, errb := testEnv(nil)
	code = run(context.Background(), e, args)
	return code, out.String(), errb.String()
}

func TestVersionReportsTheStampedVersion(t *testing.T) {
	for _, args := range [][]string{{"version"}, {"--version"}} {
		code, out, _ := runArgs(args...)
		want := "meshctl v9.8.7 " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH + "\n"
		if code != exitOK || out != want {
			t.Errorf("%v: code %d, out %q; want 0, %q", args, code, out, want)
		}
	}
}

func TestHelpListsCommands(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"-h"}, {"--help"}} {
		code, out, _ := runArgs(args...)
		if code != exitOK || !strings.Contains(out, "version") || !strings.Contains(out, "usage: meshctl") {
			t.Errorf("%v: code %d, out %q", args, code, out)
		}
	}
}

func TestUnknownCommandIsAUsageError(t *testing.T) {
	code, _, errOut := runArgs("frobnicate")
	if code != exitUsage || !strings.Contains(errOut, `unknown command "frobnicate"`) {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
}

func TestBadFlagsAreUsageErrors(t *testing.T) {
	code, _, errOut := runArgs("version", "--nope")
	if code != exitUsage || !strings.Contains(errOut, "meshctl version:") {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	code, _, _ = runArgs("version", "extra")
	if code != exitUsage {
		t.Fatalf("stray argument: code %d, want %d", code, exitUsage)
	}
}

func TestCommandHelpExitsZero(t *testing.T) {
	code, _, errOut := runArgs("version", "-h")
	if code != exitOK || !strings.Contains(errOut, "usage: meshctl version [flags]") {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
}

func TestGroupsNeedASubcommand(t *testing.T) {
	code, _, errOut := runArgs("node")
	if code != exitUsage || !strings.Contains(errOut, "missing subcommand") || !strings.Contains(errOut, "add") {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	code, _, errOut = runArgs("node", "frob")
	if code != exitUsage || !strings.Contains(errOut, `meshctl node: unknown command "frob"`) {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
}

func TestFlagsMayFollowPositionalArguments(t *testing.T) {
	// The bad name is refused after both flags parsed: proof that the
	// flags after it were read, not taken as more arguments.
	code, _, errOut := runArgs("node", "show", "Bad_Name", "--data", t.TempDir(), "--json")
	if code != exitUsage || !strings.Contains(errOut, "name:") {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	code, _, errOut = runArgs("node", "show", "a", "b")
	if code != exitUsage || !strings.Contains(errOut, "expects <name>") {
		t.Fatalf("two names: code %d, stderr %q", code, errOut)
	}
}
