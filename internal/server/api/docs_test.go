package api

import (
	"os"
	"strings"
	"testing"
)

// TestRoutesAreDocumented keeps docs/agent-protocol.md honest: every route
// this package serves appears there as "METHOD path".
func TestRoutesAreDocumented(t *testing.T) {
	b, err := os.ReadFile("../../../docs/agent-protocol.md")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	for _, r := range append(Routes(), AdminRoutes()...) {
		want := "`" + r.Method + " " + r.Path + "`"
		if !strings.Contains(doc, want) {
			t.Errorf("docs/agent-protocol.md does not list %s", want)
		}
	}
}
