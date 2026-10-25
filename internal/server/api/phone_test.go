package api

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// publicOf is the public key of the profile's private key, as the WireGuard
// app derives it.
func publicOf(t *testing.T, profile string) (priv, pub string) {
	t.Helper()
	for _, l := range strings.Split(profile, "\n") {
		if k, ok := strings.CutPrefix(l, "PrivateKey = "); ok {
			raw, err := base64.StdEncoding.DecodeString(k)
			if err != nil {
				t.Fatal(err)
			}
			sk, err := ecdh.X25519().NewPrivateKey(raw)
			if err != nil {
				t.Fatal(err)
			}
			return k, base64.StdEncoding.EncodeToString(sk.PublicKey().Bytes())
		}
	}
	t.Fatalf("no private key in the profile:\n%s", profile)
	return "", ""
}

// phoneKeyOn is the key a server's config has for the phone's peer.
func phoneKeyOn(t *testing.T, h *harness, server string) string {
	t.Helper()
	var d NodeDetail
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/"+server, nil, &d); code != 200 {
		t.Fatalf("show %s: %d %s", server, code, eb.Message)
	}
	for _, p := range d.Config.Peers {
		if p.Name == "phone" {
			return p.PublicKey
		}
	}
	return ""
}

// A phone is added without a token; each profile has a new key pair whose
// public half every server peers with, shown once and never stored or
// audited; the node says when the profile it holds is out of date, and when
// a server last saw it.
func TestPhoneProfile(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/vps-2", NodeSetRequest{CanExit: ptr(true)}, nil); code != 200 {
		t.Fatalf("vps-2 --can-exit: %d %s", code, eb.Message)
	}
	var ch NodeChange
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "phone", Role: "phone", Hub: "vps-1"}, &ch); code != 200 {
		t.Fatalf("add: %d %s", code, eb.Message)
	}
	if ch.Enrollment != nil || ch.Node.Enrolled || ch.Node.MeshIP != testfix.Phone.MeshIP {
		t.Fatalf("a phone gets no join token: %+v", ch)
	}
	before := h.latestVersion(t)

	var p PhoneProfile
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes+"/phone/phone", PhoneRequest{}, &p); code != 200 {
		t.Fatalf("profile: %d %s", code, eb.Message)
	}
	priv, pub := publicOf(t, p.Profile)
	if !strings.Contains(p.Profile, "DNS = "+testfix.VPS1.MeshIP+"\n") || !strings.Contains(p.Profile, "AllowedIPs = "+testfix.MeshCIDR+"\n") ||
		strings.Contains(p.Profile, "0.0.0.0/0") || strings.Count(p.Profile, "[Peer]") != 3 {
		t.Fatalf("mesh-only profile:\n%s", p.Profile)
	}
	if p.Version <= before || !p.Node.Enrolled || !p.Node.InSync {
		t.Fatalf("after the profile: v%d (was %d), %+v", p.Version, before, p.Node)
	}
	for _, s := range []string{"vps-1", "vps-2", "vps-3"} {
		if got := phoneKeyOn(t, h, s); got != pub {
			t.Fatalf("%s peers with the phone as %q, want the profile's key %q", s, got, pub)
		}
	}

	// A via-exit profile replaces the key: the first one stops working.
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes+"/phone/phone", PhoneRequest{Exit: "vps-2"}, &p); code != 200 {
		t.Fatalf("via-exit profile: %d %s", code, eb.Message)
	}
	priv2, pub2 := publicOf(t, p.Profile)
	if pub2 == pub || phoneKeyOn(t, h, "vps-1") != pub2 || p.Exit != "vps-2" ||
		!strings.Contains(p.Profile, "AllowedIPs = "+testfix.VPS2.MeshIP+"/32, 0.0.0.0/0, ::/0\n") {
		t.Fatalf("via-exit profile (exit %q):\n%s", p.Exit, p.Profile)
	}

	var audit []store.AuditEntry
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		audit, err = tx.AuditLog(100, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	profiles := 0
	for _, e := range audit {
		if strings.Contains(e.Detail, priv) || strings.Contains(e.Detail, priv2) {
			t.Fatalf("audit row %+v carries a private key", e)
		}
		if e.Action == "phone.profile" {
			profiles++
		}
	}
	if profiles != 2 {
		t.Fatalf("%d phone.profile audit rows, want 2", profiles)
	}

	// Its hub moves: the profile it holds is out of date.
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/phone", NodeSetRequest{Hub: ptr("vps-3")}, &ch); code != 200 || ch.Node.InSync {
		t.Fatalf("hub change: %d %s, %+v (want the profile out of date)", code, eb.Message, ch.Node)
	}

	// What the servers see of it.
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		return tx.RecordStatus(testfix.VPS1.ID, store.StatusUpdate{Peers: []store.PeerStat{{PeerID: ch.Node.ID, LastHandshake: 1234}}})
	}); err != nil {
		t.Fatal(err)
	}
	var l NodeList
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes, nil, &l); code != 200 {
		t.Fatal(code)
	}
	for _, n := range l.Nodes {
		if (n.Name == "phone") != (n.LastHandshake == 1234) {
			t.Fatalf("%s: last handshake %d", n.Name, n.LastHandshake)
		}
	}
	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/phone", nil, &d); code != 200 || d.Node.LastHandshake != 1234 {
		t.Fatalf("show: %d %+v", code, d.Node)
	}
}

// Without a server to dial, a phone gets no profile, and its key stays as it
// was.
func TestPhoneProfileNeedsAServer(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, false), storeNode(testfix.Phone, false))
	code, eb := adminCall(t, h, http.MethodPost, AdminNodes+"/phone/phone", PhoneRequest{}, nil)
	if code != http.StatusConflict || !strings.Contains(eb.Message, "join one first") {
		t.Fatalf("%d %q", code, eb.Message)
	}
	if n := h.node(t, "phone"); n.Enrolled || n.WGPubKey != "" || h.audits(t, "phone.profile") != 0 {
		t.Fatalf("a refused profile changed the phone: %+v", n)
	}
}
