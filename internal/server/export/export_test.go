package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func snapshot() store.Snapshot {
	return store.Snapshot{
		Format:   store.SnapshotFormat,
		Settings: map[string]string{store.SettingMeshCIDR: testfix.MeshCIDR, store.SettingDefaultMTU: "1420", store.SettingDNSExtraDomains: "example.lan,home.arpa"},
		Nodes: []store.SnapNode{
			{ID: 4, Name: "nas", Role: "spoke", MeshIP: testfix.NAS.MeshIP, ListenPort: 51820, PrimaryHubID: 1, Keepalive: 25, MTU: 1420,
				Enrolled: true, WGPubKey: testfix.WGKey(4), Mode: "container", ContainerSubnet: testfix.ContainerSubnet},
			{ID: 1, Name: "vps-1", Role: "server", MeshIP: testfix.VPS1.MeshIP, Endpoint: testfix.VPS1.Endpoint, ListenPort: 51820,
				CanExit: true, Keepalive: 25, MTU: 1420, Enrolled: true, WGPubKey: testfix.WGKey(1), IdentityPubKey: testfix.IdentityKeyB64(1)},
		},
		DNS:      []store.DNSRecord{{ID: 1, Name: "files", NodeID: 4}},
		Firewall: []store.FirewallRule{{ID: 1, NodeID: 1, Proto: "tcp", Port: 9000, SourceCIDR: "10.77.1.0/24", Note: "nas backup"}},
	}
}

func TestExportIsDeterministicAndByName(t *testing.T) {
	d := FromSnapshot(snapshot())
	if d.Nodes[0].Name != "vps-1" || d.Nodes[1].Hub != "vps-1" || d.DNS[0].Node != "nas" || d.Firewall[0].Node != "vps-1" {
		t.Fatalf("document %+v", d)
	}
	var a, b bytes.Buffer
	if err := Write(&a, d); err != nil {
		t.Fatal(err)
	}
	if err := Write(&b, FromSnapshot(snapshot())); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.Bytes(), b.Bytes()) {
		t.Fatal("two exports of the same state differ")
	}
	out := a.String()
	for _, want := range []string{"format: meshctl-export/1", "hub: vps-1", "can_exit: true", "source_cidr: 10.77.1.0/24",
		"mode: container", "container_subnet: " + testfix.ContainerSubnet} {
		if !strings.Contains(out, want) {
			t.Errorf("export lacks %q:\n%s", want, out)
		}
	}
	_, body, _ := strings.Cut(out, "\nformat:") // past the header comment, which names what is left out
	for _, never := range []string{"token", "private", "password", "totp"} {
		if strings.Contains(strings.ToLower(body), never) {
			t.Errorf("export mentions %q:\n%s", never, body)
		}
	}
}

func TestReadRoundTripsAndRefusesUnknownShapes(t *testing.T) {
	var b bytes.Buffer
	if err := Write(&b, FromSnapshot(snapshot())); err != nil {
		t.Fatal(err)
	}
	d, err := Read(&b)
	if err != nil || len(d.Nodes) != 2 || d.Nodes[1].WGPublicKey != testfix.WGKey(4) {
		t.Fatalf("Read = %+v, %v", d, err)
	}
	for name, in := range map[string]string{
		"unknown field":  "format: meshctl-export/1\nusers: []\n",
		"other format":   "format: meshctl-export/9\n",
		"not a document": "- 1\n- 2\n",
	} {
		if _, err := Read(strings.NewReader(in)); err == nil {
			t.Errorf("%s: Read accepted it", name)
		}
	}
}

func TestToSnapshotRoundTrips(t *testing.T) {
	in := snapshot()
	got, err := ToSnapshot(FromSnapshot(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Nodes) != 2 || got.Nodes[1].PrimaryHubID != 1 || got.DNS[0].NodeID != 4 || got.Firewall[0].NodeID != 1 ||
		got.Settings[store.SettingMeshCIDR] != testfix.MeshCIDR || got.Nodes[1].Mode != "container" ||
		got.Settings[store.SettingDNSExtraDomains] != "example.lan,home.arpa" ||
		got.Nodes[1].ContainerSubnet != testfix.ContainerSubnet || got.Nodes[0].Mode != "" {
		t.Fatalf("ToSnapshot = %+v", got)
	}
}

func TestToSnapshotRefusesHostileDocuments(t *testing.T) {
	cases := map[string]func(*Document){
		"newline in a name":       func(d *Document) { d.Nodes[0].Name = "vps-1\nrole: server" },
		"duplicate id":            func(d *Document) { d.Nodes[1].ID = d.Nodes[0].ID },
		"duplicate name":          func(d *Document) { d.Nodes[1].Name = d.Nodes[0].Name },
		"unknown role":            func(d *Document) { d.Nodes[0].Role = "router" },
		"bad mesh ip":             func(d *Document) { d.Nodes[0].MeshIP = "10.77.0.256" },
		"endpoint with a host":    func(d *Document) { d.Nodes[0].Endpoint = "vps.example:51820" },
		"listen port 0":           func(d *Document) { d.Nodes[0].ListenPort = 0 },
		"mtu out of range":        func(d *Document) { d.Nodes[0].MTU = 9000 },
		"short key":               func(d *Document) { d.Nodes[0].WGPublicKey = "abc" },
		"unknown os":              func(d *Document) { d.Nodes[0].OS = "plan9" },
		"bad hostname":            func(d *Document) { d.Nodes[0].Hostname = "a b" },
		"unknown mode":            func(d *Document) { d.Nodes[1].Mode = "bridge" },
		"bad container subnet":    func(d *Document) { d.Nodes[1].ContainerSubnet = "10.77.128.0/24\nrole: server" },
		"hub that does not exist": func(d *Document) { d.Nodes[1].Hub = "ghost" },
		"unknown setting":         func(d *Document) { d.Settings["server_url"] = "https://evil.example" },
		"bad mesh cidr":           func(d *Document) { d.Settings[store.SettingMeshCIDR] = "10.77.0.1/16" },
		"bad upstream":            func(d *Document) { d.Settings[store.SettingUpstreamDNS] = "1.1.1.1,dns.example" },
		"default hub not a node":  func(d *Document) { d.Settings[store.SettingDefaultHub] = "99" },
		"bad extra domain":        func(d *Document) { d.Settings[store.SettingDNSExtraDomains] = "lan,10.in-addr.arpa" },
		"record with both":        func(d *Document) { d.DNS[0].IP = "10.77.0.9" },
		"record name injection":   func(d *Document) { d.DNS[0].Name = "a;b" },
		"rule with bad source":    func(d *Document) { d.Firewall[0].Source = `10.0.0.0/8" accept` },
		"rule note with newline":  func(d *Document) { d.Firewall[0].Note = "a\nb" },
	}
	for name, edit := range cases {
		d := FromSnapshot(snapshot())
		edit(&d)
		if _, err := ToSnapshot(d); err == nil || !strings.Contains(err.Error(), "nothing was imported") {
			t.Errorf("%s: ToSnapshot = %v", name, err)
		}
	}
}
