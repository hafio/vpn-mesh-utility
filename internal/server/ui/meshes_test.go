package ui

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Each mesh's pages are under /m/<mesh>/ with its own nodes; "/" opens the
// first mesh; the bar names the mesh and, with several, a switcher lists
// them; the Meshes page adds, renames and removes them (amendment A49).
func TestMeshPages(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		return r.do(http.MethodGet, path, cookie, nil)
	}
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	for path, want := range map[string]string{"/": "/m/default/", "/m/default": "/m/default/"} {
		if w := get(path); w.Code != http.StatusSeeOther || w.Header().Get("Location") != want {
			t.Fatalf("GET %s = %d %s, want %s", path, w.Code, w.Header().Get("Location"), want)
		}
	}
	if w := get("/m/Bad!"); w.Code != http.StatusBadRequest {
		t.Fatalf("a bad mesh name = %d", w.Code)
	}
	body := get("/m/default/").Body.String()
	if !strings.Contains(body, `<span class="mesh">default</span>`) || strings.Contains(body, `class="meshes"`) ||
		!strings.Contains(body, `href="/m/default/nodes/vps-1"`) || !strings.Contains(body, `hx-get="/m/default/fragments/nodes"`) {
		t.Fatalf("the first mesh's dashboard:\n%s", body)
	}

	w := post("/meshes", url.Values{"name": {"lab"}, "ipv6": {"off"}})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Added mesh lab on 10.78.0.0/16, zone lab") ||
		!strings.Contains(w.Body.String(), `<a href="/m/lab/">lab</a>`) {
		t.Fatalf("add a mesh = %d\n%s", w.Code, w.Body)
	}
	w = post("/meshes", url.Values{"name": {"twin"}, "dns_zone": {"mesh"}})
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "mesh default&#39;s mesh") || !strings.Contains(w.Body.String(), `value="twin"`) {
		t.Fatalf("a mesh on another's zone = %d\n%s", w.Code, w.Body)
	}

	// Two meshes: a switcher, and each mesh's pages show its own nodes.
	body = get("/m/lab/").Body.String()
	if !strings.Contains(body, `<nav class="meshes" aria-label="Meshes">`) || !strings.Contains(body, `<a href="/m/default/">default</a>`) ||
		!strings.Contains(body, `<a href="/m/lab/" class="active" aria-current="page">lab</a>`) || strings.Contains(body, "/nodes/vps-1") {
		t.Fatalf("the lab's dashboard:\n%s", body)
	}
	if w := post("/m/lab/nodes", url.Values{"name": {"nas"}, "role": {"spoke"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `href="/m/lab/nodes/nas"`) {
		t.Fatalf("add a node to the lab = %d\n%s", w.Code, w.Body)
	}
	if len(r.nodes()) != 2 {
		t.Fatalf("the first mesh has %d nodes, want its 2", len(r.nodes()))
	}
	if w := get("/m/lab/nodes/vps-1"); w.Code != http.StatusNotFound {
		t.Fatalf("the first mesh's node in the lab = %d", w.Code)
	}
	if w := get("/m/none/nodes"); w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "no such mesh") {
		t.Fatalf("an unknown mesh = %d\n%s", w.Code, w.Body)
	}

	if w := post("/meshes/lab", url.Values{"name": {"office"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Renamed the mesh to office") {
		t.Fatalf("rename = %d\n%s", w.Code, w.Body)
	}
	if w := post("/meshes/office/remove", url.Values{}); w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "still has 1 node(s)") {
		t.Fatalf("remove a mesh with a node = %d\n%s", w.Code, w.Body)
	}
	if w := post("/meshes/default", url.Values{"name": {"office"}}); w.Code != http.StatusConflict {
		t.Fatalf("rename onto a taken name = %d", w.Code)
	}
	if w := post("/m/office/nodes/nas/remove", url.Values{}); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/m/office/nodes" {
		t.Fatalf("remove the lab's node = %d %s", w.Code, w.Header().Get("Location"))
	}
	if w := post("/meshes/office/remove", url.Values{}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Removed mesh office") {
		t.Fatalf("remove the mesh = %d\n%s", w.Code, w.Body)
	}
	if w := get("/meshes"); w.Code != http.StatusOK || strings.Contains(w.Body.String(), "office") || !strings.Contains(w.Body.String(), `<a href="/m/default/">default</a>`) {
		t.Fatalf("the Meshes page = %d\n%s", w.Code, w.Body)
	}
}
