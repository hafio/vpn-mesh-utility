package ui

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/boombuler/barcode/qr"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// The path draws exactly the dark modules of the symbol the library made.
func TestQRCode(t *testing.T) {
	text := "[Interface]\nPrivateKey = " + testfix.WGKey(9) + "\n"
	got, err := newQR(text)
	if err != nil {
		t.Fatal(err)
	}
	want, err := qr.Encode(text, qr.M, qr.Unicode)
	if err != nil {
		t.Fatal(err)
	}
	n := want.Bounds().Dx()
	if got.View != fmt.Sprintf("-4 -4 %d %d", n+8, n+8) {
		t.Fatalf("viewBox %q for a %d-module symbol", got.View, n)
	}
	drawn := map[[2]int]bool{}
	for _, cmd := range strings.Split(strings.TrimSuffix(got.Path, "z"), "z") {
		var x, y, run, back int
		if _, err := fmt.Sscanf(cmd, "M%d %dh%dv1h-%d", &x, &y, &run, &back); err != nil || run != back || run < 1 {
			t.Fatalf("path command %q: %v", cmd, err)
		}
		for i := 0; i < run; i++ {
			drawn[[2]int{x + i, y}] = true
		}
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			r, _, _, _ := want.At(x, y).RGBA()
			if dark := r < 0x8000; dark != drawn[[2]int{x, y}] {
				t.Fatalf("module (%d,%d): drawn %v, dark %v", x, y, drawn[[2]int{x, y}], dark)
			}
		}
	}
	// The finder pattern's top row: seven dark modules.
	if !strings.HasPrefix(got.Path, "M0 0h7v1h-7z") {
		t.Fatalf("path starts %q", got.Path[:20])
	}

	// Too long for any QR code: the error must not quote the content.
	secret := strings.Repeat("k", 4000)
	if _, err := newQR(secret); err == nil || strings.Contains(err.Error(), "kkkk") {
		t.Fatalf("newQR(4000 bytes) = %v", err)
	}
}

// A phone is added from the nodes page, gets its profile on its own page --
// once, as a QR code and as text, never in the log -- and its page says when
// that profile is out of date.
func TestPhonePage(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) (int, string) {
		t.Helper()
		form.Set("csrf", csrf)
		w := r.do(http.MethodPost, path, cookie, form)
		return w.Code, w.Body.String()
	}
	if code, body := post("/m/default/nodes/vps-2", url.Values{"endpoint": {testfix.VPS2.Endpoint}, "listen_port": {"51820"}, "mtu": {"1420"}, "can_exit": {"on"}}); code != http.StatusOK {
		t.Fatalf("vps-2 can exit: %d\n%s", code, body)
	}

	code, body := post("/m/default/nodes", url.Values{"name": {"phone"}, "role": {"phone"}, "hub": {"vps-1"}})
	if code != http.StatusOK || !strings.Contains(body, "Phone phone added") || !strings.Contains(body, "Make a profile") ||
		!strings.Contains(body, `<option value="vps-2">through vps-2</option>`) || !strings.Contains(body, "no profile yet") {
		t.Fatalf("add a phone = %d\n%s", code, body)
	}
	for _, absent := range []string{"New join token", "Run in a container", "Peers as last reported", `<option value="">off</option>`} {
		if strings.Contains(body, absent) {
			t.Fatalf("a phone's page has %q:\n%s", absent, body)
		}
	}

	r.logs.Reset()
	w := r.do(http.MethodPost, "/m/default/nodes/phone/phone", cookie, url.Values{"csrf": {csrf}, "exit": {"vps-2"}})
	body = w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `<svg class="qr" viewBox="-4 -4 `) || !strings.Contains(body, "PrivateKey = ") ||
		!strings.Contains(body, "all of the phone's traffic goes through vps-2") || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("profile = %d %v\n%s", w.Code, w.Header(), body)
	}
	// html/template escapes the key's '+' as &#43;; the browser shows '+'.
	_, key, _ := strings.Cut(body, "PrivateKey = ")
	key, _, _ = strings.Cut(key, "\n")
	if key = html.UnescapeString(key); len(key) != 44 || strings.Contains(r.logs.String(), key) || strings.Contains(r.logs.String(), "PrivateKey") {
		t.Fatalf("the private key %q is in the log, or not in the page: %s", key, r.logs.String())
	}
	if w := r.do(http.MethodGet, "/m/default/nodes/phone", cookie, nil); !strings.Contains(w.Body.String(), "no handshake yet") ||
		!strings.Contains(w.Body.String(), "current (config v") {
		t.Fatalf("after the profile:\n%s", w.Body)
	}

	if code, body := post("/m/default/nodes/phone/phone", url.Values{"exit": {"vps-1"}}); code != http.StatusConflict || !strings.Contains(body, "cannot be this phone&#39;s exit") ||
		strings.Contains(body, "PrivateKey") {
		t.Fatalf("via a server that is no exit = %d\n%s", code, body)
	}
	if code, body := post("/m/default/nodes/phone", url.Values{"hub": {"vps-2"}, "keepalive": {"25"}, "mtu": {"1420"}}); code != http.StatusOK ||
		!strings.Contains(body, "profile out of date") || !strings.Contains(body, "make a new one") {
		t.Fatalf("hub change = %d\n%s", code, body)
	}

	var audit []store.AuditEntry
	if err := r.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		audit, err = tx.AuditLog(10, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	profiles := 0
	for _, e := range audit {
		if e.Action == "phone.profile" && e.Actor == "user:admin" {
			profiles++
		}
	}
	if profiles != 1 {
		t.Fatalf("%d phone.profile rows by user:admin in %+v", profiles, audit)
	}
}
