package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// An admin on a control plane with no mesh is sent to the Meshes page.
func TestHomeWithoutMeshes(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.signIn()
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.DeleteMesh(store.FirstMesh) }); err != nil {
		t.Fatal(err)
	}
	if w := r.do(http.MethodGet, "/", cookie, nil); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/meshes" {
		t.Fatalf("home without a mesh = %d %s", w.Code, w.Header().Get("Location"))
	}
}

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

// A user who is not an admin manages only the meshes granted to them
// (amendment A49): the others are the same 404 as no mesh, the switcher
// lists only theirs, the Meshes and Users pages are an admin's, and their
// audit log holds their mesh's entries alone. The Users page grants,
// revokes and changes roles.
func TestAccessByGrant(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	if w := post("/meshes", url.Values{"name": {"lab"}, "ipv6": {"off"}}); w.Code != http.StatusOK {
		t.Fatalf("add the lab = %d\n%s", w.Code, w.Body)
	}
	w := post("/users", url.Values{"username": {"ops"}, "password": {password}, "password2": {password}})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Added ops; grant them meshes below") || !strings.Contains(w.Body.String(), "no mesh yet") {
		t.Fatalf("add ops = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users", url.Values{"username": {"boss"}, "password": {password}, "password2": {password}, "admin": {"on"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Added boss as an admin") || !strings.Contains(w.Body.String(), "<option>lab</option>") {
		t.Fatalf("add boss = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users", url.Values{"username": {"carol"}, "password": {password}, "password2": {password}, "mesh": {"lab"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Added carol, who manages mesh lab") {
		t.Fatalf("add carol for the lab = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users", url.Values{"username": {"dave"}, "password": {password}, "password2": {password}, "mesh": {"lab"}, "admin": {"on"}}); w.Code != http.StatusBadRequest {
		t.Fatalf("an admin for one mesh = %d", w.Code)
	}

	// ops has no mesh yet.
	opsCookie, _ := r.signInAs("ops")
	opsGet := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		return r.do(http.MethodGet, path, opsCookie, nil)
	}
	if w := opsGet("/"); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "ask an admin") || strings.Contains(w.Body.String(), `href="/users"`) {
		t.Fatalf("ops without a mesh = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users/ops/grant", url.Values{"mesh": {"lab"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Granted mesh lab to ops") {
		t.Fatalf("grant = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users/ops/grant", url.Values{"mesh": {"none"}}); w.Code != http.StatusNotFound {
		t.Fatalf("grant of no mesh = %d", w.Code)
	}

	// ops manages the lab alone.
	if w := opsGet("/"); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/m/lab/" {
		t.Fatalf("ops's home = %d %s", w.Code, w.Header().Get("Location"))
	}
	body := opsGet("/m/lab/").Body.String()
	if strings.Contains(body, `class="meshes"`) || strings.Contains(body, `href="/meshes"`) || strings.Contains(body, `href="/users"`) {
		t.Fatalf("ops's dashboard shows another mesh or an admin page:\n%s", body)
	}
	for _, path := range []string{"/m/default/", "/m/default/nodes/vps-1", "/m/default/export"} {
		if w := opsGet(path); w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "you manage no such mesh") {
			t.Fatalf("ops at %s = %d", path, w.Code)
		}
	}
	for _, path := range []string{"/meshes", "/users"} {
		if w := opsGet(path); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "only an admin") {
			t.Fatalf("ops at %s = %d", path, w.Code)
		}
	}
	if body := opsGet("/m/lab/audit").Body.String(); !strings.Contains(body, "user.grant") || strings.Contains(body, "user.add") {
		t.Fatalf("ops's audit log:\n%s", body)
	}
	if body := r.do(http.MethodGet, "/m/lab/audit", cookie, nil).Body.String(); !strings.Contains(body, "user.add") {
		t.Fatalf("the admin's audit log lacks the control plane's entries:\n%s", body)
	}

	// Roles: ops becomes an admin and back; the admin's own role stays.
	if w := post("/users/ops/admin", url.Values{"admin": {"on"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Made an admin: ops") {
		t.Fatalf("make ops an admin = %d\n%s", w.Code, w.Body)
	}
	if w := opsGet("/m/default/"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `class="meshes"`) {
		t.Fatalf("ops as an admin = %d", w.Code)
	}
	if w := post("/users/ops/admin", url.Values{"admin": {"off"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "No longer an admin: ops") {
		t.Fatalf("take it away = %d\n%s", w.Code, w.Body)
	}
	if w := post("/users/admin/admin", url.Values{"admin": {"off"}}); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "your own role") {
		t.Fatalf("an admin's own role = %d", w.Code)
	}
	if w := post("/users/ops/revoke", url.Values{"mesh": {"lab"}}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Took mesh lab away from ops") {
		t.Fatalf("revoke = %d\n%s", w.Code, w.Body)
	}
	if w := opsGet("/m/lab/"); w.Code != http.StatusNotFound {
		t.Fatalf("ops after the revoke = %d", w.Code)
	}
}
