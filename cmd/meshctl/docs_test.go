package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
				fmt.Fprintf(&b, "| `--%s` | %s | %s |\n", f.Name, def, cell(f.Usage))
			})
		}
	}
	walk("meshctl", commands())
	return b.String()
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
			if code, _, errOut := runArgs(append(args, "-h")...); code != exitOK || !strings.Contains(errOut, "usage: meshctl") {
				t.Errorf("%v -h: code %d, stderr %q", args, code, errOut)
			}
		}
	}
	walk(nil, commands())
}
