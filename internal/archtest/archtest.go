// Package archtest enforces the module boundaries the design depends on
// (CLAUDE.md "Packages, and which way they may import"): the server and the
// agent never import each other, the shared packages import neither, and each
// side stays clear of the other side's heavy dependencies. The check reads
// import lists with go/build for every shipped GOOS, without the network or
// module resolution, so it runs in `go test ./...` on both CI legs.
package archtest

import (
	"errors"
	"fmt"
	"go/build"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Rule forbids packages under Dir (a path relative to the module root, using
// forward slashes) from importing any path that starts with one of Deny.
// Entries in Deny that begin with "./" are module-relative. Packages under
// one of Except are exempt (the one sanctioned user of a denied import).
type Rule struct {
	Dir    string
	Deny   []string
	Except []string
	// ProdOnly limits the rule to non-test files (for "production code must
	// not import the test fixtures").
	ProdOnly bool
	Why      string
}

// Violation is one forbidden import.
type Violation struct {
	Package string
	Import  string
	GOOS    string
	Why     string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s imports %s (GOOS=%s): %s", v.Package, v.Import, v.GOOS, v.Why)
}

// GOOSes are the platforms meshctl ships for; build-tagged files of each are
// checked.
var GOOSes = []string{"linux", "windows"}

// Check walks every package directory under root (skipping testdata, hidden
// and underscore directories) and reports each import that a rule forbids.
// Test files (internal and external) count: a test that imports across the
// boundary couples the two sides just as surely.
func Check(root, module string, rules []Rule) ([]Violation, error) {
	var out []Violation
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if p != root && (name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		for _, goos := range GOOSes {
			prod, all, err := importsOf(p, goos)
			if err != nil {
				return err
			}
			for _, r := range rules {
				if !under(rel, r.Dir) || underAny(rel, r.Except) {
					continue
				}
				imports := all
				if r.ProdOnly {
					imports = prod
				}
				for _, imp := range imports {
					for _, deny := range r.Deny {
						target := deny
						if strings.HasPrefix(deny, "./") {
							target = module + "/" + strings.TrimPrefix(deny, "./")
						}
						if imp == target || strings.HasPrefix(imp, target+"/") {
							v := Violation{Package: module + "/" + rel, Import: imp, GOOS: goos, Why: r.Why}
							if !seen[v.String()] {
								seen[v.String()] = true
								out = append(out, v)
							}
						}
					}
				}
			}
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, err
}

// under reports whether rel is dir or inside it.
func under(rel, dir string) bool {
	return rel == dir || strings.HasPrefix(rel, dir+"/")
}

func underAny(rel string, dirs []string) bool {
	for _, d := range dirs {
		if under(rel, d) {
			return true
		}
	}
	return false
}

// importsOf returns the production imports and all imports (production plus
// test) of the package in dir as built for goos. A directory without Go files
// yields nothing.
func importsOf(dir, goos string) (prod, all []string, err error) {
	ctx := build.Default
	ctx.GOOS = goos
	ctx.GOARCH = "amd64"
	ctx.CgoEnabled = false
	pkg, err := ctx.ImportDir(dir, build.ImportComment)
	if err != nil {
		var noGo *build.NoGoError
		if errors.As(err, &noGo) {
			return nil, nil, nil
		}
		// A directory whose files are all excluded for this GOOS is fine too.
		if pkg != nil && len(pkg.GoFiles)+len(pkg.TestGoFiles)+len(pkg.XTestGoFiles) == 0 {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("read imports of %s for GOOS=%s: %w", dir, goos, err)
	}
	prod = append([]string{}, pkg.Imports...)
	all = append(append(append([]string{}, pkg.Imports...), pkg.TestImports...), pkg.XTestImports...)
	return prod, all, nil
}

// ModuleRoot walks up from dir to the directory holding go.mod and returns it
// with the module path declared there.
func ModuleRoot(dir string) (root, module string, err error) {
	for d := dir; ; d = filepath.Dir(d) {
		b, err := os.ReadFile(filepath.Join(d, "go.mod"))
		if err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					return d, strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
				}
			}
			return "", "", fmt.Errorf("%s/go.mod has no module line", d)
		}
		if filepath.Dir(d) == d {
			return "", "", fmt.Errorf("no go.mod above %s", dir)
		}
	}
}
