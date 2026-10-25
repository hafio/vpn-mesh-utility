package ui

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// The Firewall pages build a mesh firewall one item at a time: objects, a
// zone, a policy; move, rename and check it; an item still named elsewhere
// stays; every change needs the session's CSRF token.
func TestFirewallPages(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	page := func(path string, wants ...string) string {
		t.Helper()
		w := r.do(http.MethodGet, path, cookie, nil)
		for _, want := range wants {
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), want) {
				t.Fatalf("GET %s = %d, lacks %q:\n%s", path, w.Code, want, w.Body)
			}
		}
		return w.Body.String()
	}
	ok := func(w *httptest.ResponseRecorder, wants ...string) {
		t.Helper()
		for _, want := range wants {
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), want) {
				t.Fatalf("= %d, lacks %q:\n%s", w.Code, want, w.Body)
			}
		}
	}
	refusedWith := func(w *httptest.ResponseRecorder, status int, wants ...string) {
		t.Helper()
		for _, want := range wants {
			if w.Code != status || !strings.Contains(w.Body.String(), want) {
				t.Fatalf("= %d (want %d), lacks %q:\n%s", w.Code, status, want, w.Body)
			}
		}
	}

	page("/m/default/firewall", "allow-all", "system: ping", "drop everything else", `href="/m/default/firewall/zones"`, `name="position" value="1"`)
	ok(post("/m/default/firewall/objects/group", url.Values{"name": {"core"}, "entries": {"vps-1"}}), "Group core saved")
	ok(post("/m/default/firewall/objects/service", url.Values{"name": {"web"}, "entries": {"tcp/80, tcp/443"}}), "Service web saved")
	ok(post("/m/default/firewall/objects/address", url.Values{"name": {"lan"}, "entries": {"10.77.5.0/24"}}), "Address lan saved")
	page("/m/default/firewall/objects", "<code>group:core</code>", "<td>vps-1</td>", "tcp/80, tcp/443", "10.77.5.0/24")
	ok(post("/m/default/firewall/zones", url.Values{"name": {"office"}, "intra": {"allow"}, "members": {"group:core addr:lan"}}), "Zone office saved")
	ok(post("/m/default/firewall/policies", url.Values{"name": {"web"}, "action": {"allow"}, "from": {"zone:office"}, "to": {"node:vps-2"},
		"services": {"svc:web"}, "position": {"1"}}), "Policy web saved", "(policy: web)")
	ok(post("/m/default/firewall/policies/web/move", url.Values{"position": {"2"}}), "Policy web moved to 2")
	page("/m/default/firewall/policies/web", `value="web"`, `value="zone:office"`, "Policy 2 of 2")
	ok(post("/m/default/firewall/policies/web", url.Values{"name": {"web2"}, "action": {"allow"}, "from": {"zone:office"}, "to": {"node:vps-2"},
		"services": {"svc:web"}, "position": {"2"}}), "Policy web2 saved")
	page("/m/default/firewall/zones/office", "<h2>Outbound</h2>", "(policy: web2)", "10.77.0.1, 10.77.5.0/24")
	page("/m/default/firewall/zones", `href="/m/default/firewall/zones/office"`, "10.77.0.1, 10.77.5.0/24")
	page("/m/default/firewall/reach?from=vps-1&to=vps-2&service=tcp/80", "<strong>allowed</strong>", "rule 4: policy: allow-all", "<th>office</th>")

	// Schedules: days and times on the form, read in the zone set on the
	// page (the clock reads Monday 08:00 UTC, 16:00 in Singapore).
	ok(post("/m/default/firewall/timezone", url.Values{"timezone": {"Asia/Singapore"}}), "Schedules now follow Asia/Singapore time",
		`value="Asia/Singapore"`)
	ok(post("/m/default/firewall/policies", url.Values{"name": {"nights"}, "action": {"deny"}, "from": {"zone:office"}, "days": {"sat", "sun"},
		"start": {"22:00"}, "end": {"06:00"}, "position": {"1"}}), "Policy nights saved", "sat,sun 22:00-06:00", "(not now)")
	page("/m/default/firewall/policies/nights", `value="sat" checked`, `value="mon">`, `value="22:00"`, "Times are in Asia/Singapore")
	page("/m/default/firewall/reach?from=vps-1&to=vps-2&service=tcp/80&at=2026-10-17T23:00:00%2B08:00", "<strong>dropped</strong>",
		"at 2026-10-17T23:00:00&#43;08:00", "rule 4: policy: nights") // html/template writes + as &#43;
	page("/m/default/firewall/reach?from=vps-1&to=vps-2&service=tcp/80", "Not in force then (scheduled): policy: nights")
	refusedWith(post("/m/default/firewall/policies", url.Values{"name": {"x"}, "action": {"allow"}, "start": {"9am"}, "end": {"18:00"}}),
		http.StatusBadRequest, "schedule.start", `value="9am"`)
	refusedWith(post("/m/default/firewall/timezone", url.Values{"timezone": {"Mars/Olympus"}}), http.StatusBadRequest, "timezone")
	ok(post("/m/default/firewall/policies/nights/remove", url.Values{}), "Policy nights removed")
	ok(post("/m/default/firewall/timezone", url.Values{"timezone": {""}}), "Schedules now follow the server&#39;s time", `placeholder="the server's: UTC"`)

	// Refusals keep the form and change nothing.
	refusedWith(post("/m/default/firewall/zones/office/remove", url.Values{}), http.StatusConflict, `there is no zone &#34;office&#34;`)
	refusedWith(post("/m/default/firewall/objects/service/web/remove", url.Values{}), http.StatusConflict, `there is no service object &#34;web&#34;`)
	refusedWith(post("/m/default/firewall/policies", url.Values{"name": {"bad"}, "action": {"accept"}}), http.StatusBadRequest, "action", `value="bad"`)
	refusedWith(post("/m/default/firewall/policies", url.Values{"name": {"x"}, "action": {"allow"}, "position": {"first"}}), http.StatusBadRequest, "position")
	refusedWith(post("/m/default/firewall/zones/office", url.Values{"name": {"office"}, "intra": {"allow"}, "members": {"node:ghost"}}),
		http.StatusConflict, `there is no node &#34;ghost&#34;`, `value="node:ghost"`)
	refusedWith(post("/m/default/firewall/objects/address", url.Values{"name": {"lan"}, "entries": {"10.77.5.1/24"}}), http.StatusBadRequest,
		"host bits", `value="10.77.5.1/24"`)
	refusedWith(post("/m/default/firewall/objects/rule", url.Values{"name": {"x"}}), http.StatusBadRequest, "group, address or service")
	refusedWith(post("/m/default/firewall/policies/ghost/move", url.Values{"position": {"1"}}), http.StatusNotFound, `there is no policy &#34;ghost&#34;`)
	refusedWith(r.do(http.MethodGet, "/m/default/firewall/policies/ghost", cookie, nil), http.StatusNotFound, "there is no policy")
	refusedWith(r.do(http.MethodGet, "/m/default/firewall/zones/ghost", cookie, nil), http.StatusNotFound, "there is no zone")
	refusedWith(r.do(http.MethodGet, "/m/default/firewall/reach?from=vps-1&to=vps-2&service=tcp", cookie, nil), http.StatusBadRequest, "check one port")
	if w := r.do(http.MethodPost, "/m/default/firewall/policies/web2/remove", cookie, url.Values{}); w.Code != http.StatusForbidden {
		t.Fatalf("a removal without the CSRF token = %d", w.Code)
	}

	// In order, everything goes.
	ok(post("/m/default/firewall/policies/web2/remove", url.Values{}), "Policy web2 removed")
	ok(post("/m/default/firewall/zones/office/remove", url.Values{}), "Zone office removed")
	for _, kind := range []string{"service/web", "address/lan", "group/core"} {
		ok(post("/m/default/firewall/objects/"+kind+"/remove", url.Values{}), "removed")
	}
	if body := page("/m/default/firewall/objects"); strings.Contains(body, "group:core") || strings.Contains(body, "svc:web") {
		t.Fatalf("objects left:\n%s", body)
	}
	page("/m/default/firewall/zones", "No zones yet.")
	page("/m/default/firewall/reach", "No zones yet")
}

// A node's groups are set where it is added and on its page.
func TestNodeGroupsOnPages(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	form := url.Values{"csrf": {csrf}, "name": {"box"}, "role": {"spoke"}, "mode": {"host"}, "groups": {"lab, office"}}
	if w := r.do(http.MethodPost, "/m/default/nodes", cookie, form); w.Code != http.StatusOK {
		t.Fatalf("add with groups = %d\n%s", w.Code, w.Body)
	}
	if body := r.do(http.MethodGet, "/m/default/nodes/box", cookie, nil).Body.String(); !strings.Contains(body, "<th>Groups</th><td>lab, office") ||
		!strings.Contains(body, `name="groups" value="lab, office"`) {
		t.Fatalf("the node page lacks its groups:\n%s", body)
	}
	edit := url.Values{"csrf": {csrf}, "mtu": {"1420"}, "keepalive": {"25"}, "hub": {""}, "exit": {""}, "groups": {"lab"}}
	if w := r.do(http.MethodPost, "/m/default/nodes/box", cookie, edit); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `name="groups" value="lab"`) {
		t.Fatalf("edit groups = %d\n%s", w.Code, w.Body)
	}
	edit.Set("groups", "Lab")
	if w := r.do(http.MethodPost, "/m/default/nodes/box", cookie, edit); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "groups[0]") {
		t.Fatalf("a bad group = %d\n%s", w.Code, w.Body)
	}
}
