package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

// Every doc opens with a table of contents made from its ## and ###
// headings; ./scripts/dev.sh regen rewrites them (TestDocsOpenWithContents
// -update), and docs/commands.md gets its own from renderCommands.

const contentsHeading = "## Contents"

var reLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// slug is the anchor GitHub gives a heading: lower case, letters, digits,
// '-' and '_' kept, spaces made '-', everything else dropped.
func slug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(reLink.ReplaceAllString(heading, "$1")) {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			b.WriteRune(r)
		case r == ' ':
			b.WriteByte('-')
		}
	}
	return b.String()
}

// heading reports a Markdown heading's level and text; level 0 is no heading.
func heading(line string) (int, string) {
	level := len(line) - len(strings.TrimLeft(line, "#"))
	if level == 0 || level > 6 || !strings.HasPrefix(line[level:], " ") {
		return 0, ""
	}
	return level, strings.TrimSpace(line[level:])
}

// withContents returns doc with its table of contents as its first section:
// the one it has is replaced, else the contents go before the first other
// section. Headings in fenced code are not headings.
func withContents(doc string) string {
	lines := strings.Split(doc, "\n")
	var kept []string
	fenced, skipping := false, false
	for _, l := range lines {
		if strings.HasPrefix(l, "```") {
			fenced = !fenced
		}
		level, _ := heading(l)
		switch {
		case !fenced && l == contentsHeading:
			skipping = true
			continue
		case skipping && !fenced && level > 0 && level <= 2:
			skipping = false
		case skipping:
			continue
		}
		kept = append(kept, l)
	}

	// GitHub numbers repeated anchors -1, -2, ... in document order; the
	// contents' own heading comes first among the sections.
	seen := map[string]int{slug(strings.TrimPrefix(contentsHeading, "## ")): 1}
	var toc []string
	first := -1
	fenced = false
	for i, l := range kept {
		if strings.HasPrefix(l, "```") {
			fenced = !fenced
			continue
		}
		level, text := heading(l)
		if fenced || level == 0 {
			continue
		}
		anchor := slug(text)
		if n := seen[anchor]; n > 0 {
			seen[anchor] = n + 1
			anchor = fmt.Sprintf("%s-%d", anchor, n)
		} else {
			seen[anchor] = 1
		}
		if level == 2 && first < 0 {
			first = i
		}
		if level == 2 || level == 3 {
			text = strings.NewReplacer("—", "--", "–", "-").Replace(text)
			toc = append(toc, fmt.Sprintf("%s- [%s](#%s)", strings.Repeat("  ", level-2), text, anchor))
		}
	}
	if len(toc) == 0 {
		return strings.Join(kept, "\n")
	}
	block := append([]string{contentsHeading, ""}, toc...)
	block = append(block, "")
	if first < 0 {
		return strings.Join(append(kept, append([]string{""}, block...)...), "\n")
	}
	out := append(append(append([]string{}, kept[:first]...), block...), kept[first:]...)
	return strings.Join(out, "\n")
}

// docFiles are the repository's docs: README.md and docs/*.md.
func docFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../../docs/*.md")
	if err != nil || len(files) == 0 {
		t.Fatalf("no docs under ../../docs: %v", err)
	}
	return append(files, "../../README.md")
}

func TestDocsOpenWithContents(t *testing.T) {
	for _, f := range docFiles(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		want := withContents(string(b))
		if *update && want != string(b) {
			if err := os.WriteFile(f, []byte(want), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if want != string(b) {
			t.Errorf("%s: its table of contents does not match its headings; run ./scripts/dev.sh regen", filepath.Base(f))
		}
		if !strings.Contains(string(b), "\n"+contentsHeading+"\n") {
			t.Errorf("%s has no %q section", filepath.Base(f), contentsHeading)
		}
	}
}

// Contents list ## and ### headings with GitHub's anchors (repeats
// numbered), leave fenced code alone, and are rewritten in place.
func TestWithContents(t *testing.T) {
	got := withContents("# Title\n\nIntro.\n\n## One\n\n### A's part\n\n```\n## not a heading\n```\n\n## One\n")
	if want := "# Title\n\nIntro.\n\n## Contents\n\n- [One](#one)\n  - [A's part](#as-part)\n- [One](#one-1)\n\n" +
		"## One\n\n### A's part\n\n```\n## not a heading\n```\n\n## One\n"; got != want {
		t.Fatalf("withContents:\n%s\nwant:\n%s", got, want)
	}
	if again := withContents(strings.Replace(got, "- [One](#one)\n", "", 1)); again != got {
		t.Fatalf("withContents did not rewrite stale contents:\n%s", again)
	}
	for _, c := range []struct{ in, want string }{
		{"1. Run the control plane (Dokploy)", "1-run-the-control-plane-dokploy"},
		{"A host-mode node's containers", "a-host-mode-nodes-containers"},
		{"Shared fixtures -- `internal/testfix`", "shared-fixtures----internaltestfix"},
		{"P0 \u2014 must ship in v1", "p0--must-ship-in-v1"},
		{"See [the spec](spec.md)", "see-the-spec"},
	} {
		if got := slug(c.in); got != c.want {
			t.Errorf("slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
