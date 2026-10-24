package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "rewrite docs/commands.md from the command registry")

const commandsDoc = "../../docs/commands.md"

const commandsHeader = `# Command reference

Every ` + "`meshctl`" + ` command and flag. This file is generated from the command registry in
` + "`cmd/meshctl`" + ` by ` + "`TestCommandsDoc`" + `; after changing a command, run
` + "`./scripts/dev.sh regen`" + ` and commit the result. Flags may come before or after the
positional arguments, and ` + "`-flag`" + ` and ` + "`--flag`" + ` are the same. Defaults are the Linux ones.
`

// renderCommands renders the command tree as Markdown, with Linux defaults
// and an empty environment so the output is the same on every machine.
func renderCommands() string {
	e := &env{version: "dev", stdin: strings.NewReader(""), stdout: io.Discard, stderr: io.Discard,
		goos: "linux", getenv: func(string) string { return "" }, now: time.Now}
	var b strings.Builder
	b.WriteString(commandsHeader)
	var walk func(path string, list []command)
	walk = func(path string, list []command) {
		sorted := slices.Clone(list)
		slices.SortFunc(sorted, func(x, y command) int { return strings.Compare(x.name, y.name) })
		for _, c := range sorted {
			p := path + " " + c.name
			if c.setup == nil {
				walk(p, c.sub)
				continue
			}
			fs := flag.NewFlagSet(p, flag.ContinueOnError)
			c.setup(fs, e)
			fmt.Fprintf(&b, "\n## %s\n\n%s.\n\n```\n%s%s [flags]\n```\n", p, sentence(c.summary), p, argSynopsis(c.args))
			first := true
			fs.VisitAll(func(f *flag.Flag) {
				if first {
					b.WriteString("\n| Flag | Default | Meaning |\n| --- | --- | --- |\n")
					first = false
				}
				def := ""
				if f.DefValue != "" {
					def = "`" + f.DefValue + "`"
				}
				fmt.Fprintf(&b, "| `%s` | %s | %s |\n", flagName(f.Name), def, cell(f.Usage))
			})
			if len(c.env) > 0 {
				b.WriteString("\n| Variable | Mode | Default | Meaning |\n| --- | --- | --- | --- |\n")
				for _, v := range c.env {
					def, meaning := envDoc(e, v)
					if def != "" {
						def = "`" + def + "`"
					}
					mode := v.mode
					if mode == "" {
						mode = "both"
					}
					fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", v.name, mode, def, cell(meaning))
				}
			}
		}
	}
	walk("meshctl", commands())
	return b.String()
}

// envDoc is a start variable's default and meaning: its own, or for one
// handed on as a flag without one, that flag's.
func envDoc(e *env, v envVar) (def, meaning string) {
	if v.flag == "" || v.meaning != "" {
		return v.def, v.meaning
	}
	f, cmd := startFlag(e, v)
	if f == nil {
		return "", "no such flag --" + v.flag
	}
	return f.DefValue, "`" + cmd + " " + flagName(f.Name) + "`: " + f.Usage
}

// startFlag is the flag a start variable is handed on as: the first of its
// modes' commands (`control`, then `agent run`) that has it.
func startFlag(e *env, v envVar) (*flag.Flag, string) {
	for _, c := range []struct {
		mode, cmd string
		setup     func(*flag.FlagSet, *env) action
	}{{modeControl, "meshctl control", setupControl}, {modeNode, "meshctl agent run", setupAgentRun}} {
		if v.mode != "" && v.mode != c.mode {
			continue
		}
		fs := flag.NewFlagSet(c.cmd, flag.ContinueOnError)
		c.setup(fs, e)
		if f := fs.Lookup(v.flag); f != nil {
			return f, c.cmd
		}
	}
	return nil, ""
}

func sentence(s string) string { return strings.ToUpper(s[:1]) + s[1:] }

func cell(s string) string {
	return strings.NewReplacer("|", `\|`, "<", "&lt;", ">", "&gt;").Replace(s)
}

func TestCommandsDoc(t *testing.T) {
	got := renderCommands()
	if *update {
		if err := os.WriteFile(commandsDoc, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(commandsDoc)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(gl) || i < len(wl); i++ {
			var g, w string
			if i < len(gl) {
				g = gl[i]
			}
			if i < len(wl) {
				w = wl[i]
			}
			if g != w {
				t.Fatalf("%s is out of date at line %d (run ./scripts/dev.sh regen):\n got: %q\nwant: %q", filepath.Base(commandsDoc), i+1, g, w)
			}
		}
	}
}

// Every leaf command answers -h with exit 0, and every group lists its
// subcommands: the registry has no dead entries.
func TestEveryCommandHasHelp(t *testing.T) {
	var walk func(prefix []string, list []command)
	walk = func(prefix []string, list []command) {
		for _, c := range list {
			args := append(slices.Clone(prefix), c.name)
			if c.setup == nil {
				code, out, _ := runArgs(append(args, "-h")...)
				for _, s := range c.sub {
					if code != exitOK || !strings.Contains(out, s.name) {
						t.Errorf("%v -h: code %d, missing %q in:\n%s", args, code, s.name, out)
					}
				}
				walk(args, c.sub)
				continue
			}
			code, _, errOut := runArgs(append(args, "-h")...)
			if code != exitOK || !strings.Contains(errOut, "usage: meshctl") {
				t.Errorf("%v -h: code %d, stderr %q", args, code, errOut)
			}
			// Flags are spelled as the docs spell them: --name, or -f.
			for _, line := range strings.Split(errOut, "\n") {
				if f, ok := strings.CutPrefix(line, "  -"); ok && !strings.HasPrefix(f, "-") && len(strings.Fields(f + " x")[0]) > 1 {
					t.Errorf("%v -h spells a flag with one dash: %q", args, line)
				}
			}
		}
	}
	walk(nil, commands())
}

// Help shows each flag as the docs do, with its kind and its default.
func TestHelpShowsFlagsAsTheDocsDo(t *testing.T) {
	_, _, join := runArgs("agent", "join", "-h")
	_, _, run := runArgs("agent", "run", "-h")
	_, _, logs := runArgs("logs", "-h")
	for _, c := range []struct{ out, want string }{
		{join, "  --server string\n    \tthe control plane's origin, e.g. https://mesh.example.com\n"},
		{join, "  --token string\n"},
		{run, `(default "auto")`},
		{run, "  --status-every duration\n"},
		{run, "(default 30s)"},
		{logs, "  -f\n    \tkeep printing new lines until interrupted\n"},
	} {
		if !strings.Contains(c.out, c.want) {
			t.Errorf("help lacks %q:\n%s", c.want, c.out)
		}
	}
	if strings.Contains(run, "(default false)") || strings.Contains(join, `(default "")`) {
		t.Errorf("help shows a zero default:\n%s%s", run, join)
	}
}

// Every MESHCTL_ variable docs/containers.md names is one `meshctl start`
// reads (or MESHCTL_VERSION, which the control plane's compose files take
// for the image tag), so a sample cannot set one that does nothing.
func TestContainersDocNamesRealVariables(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "containers.md"))
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{"MESHCTL_VERSION": true}
	for _, v := range startVars() {
		known[v.name] = true
	}
	names := regexp.MustCompile(`MESHCTL_[A-Z0-9_]+`).FindAllString(string(b), -1)
	for _, name := range names {
		if !known[name] {
			t.Errorf("docs/containers.md names %s, which meshctl start does not read", name)
		}
	}
	if len(names) == 0 {
		t.Fatal("docs/containers.md names no variable")
	}
}
