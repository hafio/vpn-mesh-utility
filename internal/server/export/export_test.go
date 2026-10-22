package export

import (
	"bytes"
	"reflect"
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
		DNS: []store.DNSRecord{{ID: 1, Name: "files", NodeID: 4}},
		Segmentation: &store.Segmentation{
			Groups: []store.SegGroup{{Name: "storage", Nodes: []string{"nas"}}},
			Zones:  []store.SegZone{{Name: "home", Intra: store.SegAllow, Members: []string{"group:storage"}}},
			Policies: []store.SegPolicy{{Name: "backup", Description: "nas backup", From: []string{"zone:home"}, To: []string{"node:vps-1"},
				Services: []string{"tcp/9000"}, Action: store.SegAllow,
				Schedule: &store.SegSchedule{Days: []string{"sun"}, Start: "01:00", End: "05:00"}}, store.DefaultPolicy()},
			TimeZone: "Europe/Berlin",
		},
	}
}

// legacy is an export made before the mesh firewall: custom rules instead.
func legacy() Document {
	d := FromSnapshot(snapshot())
	d.Segmentation = nil
	d.Firewall = []FirewallRule{{Node: "vps-1", Proto: "tcp", Port: 9000, Source: "10.77.1.0/24", Note: "nas backup"}, {Proto: "udp", Port: 161, Source: "0.0.0.0/0"}}
	return d
}

func TestExportIsDeterministicAndByName(t *testing.T) {
	d := FromSnapshot(snapshot())
	if d.Nodes[0].Name != "vps-1" || d.Nodes[1].Hub != "vps-1" || d.DNS[0].Node != "nas" || d.Segmentation.Groups[0].Nodes[0] != "nas" {
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
	for _, want := range []string{"format: meshctl-export/1", "hub: vps-1", "can_exit: true", "segmentation:", "- group:storage",
		"- tcp/9000", "name: allow-all", "mode: container", "container_subnet: " + testfix.ContainerSubnet, "timezone: Europe/Berlin", "- sun"} {
		if !strings.Contains(out, want) {
			t.Errorf("export lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "firewall_rules") {
		t.Errorf("a new export carries the old custom rules:\n%s", out)
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
	if len(got.Nodes) != 2 || got.Nodes[1].PrimaryHubID != 1 || got.DNS[0].NodeID != 4 || !reflect.DeepEqual(got.Segmentation, in.Segmentation) ||
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
		"rule with bad source":    func(d *Document) { *d = legacy(); d.Firewall[0].Source = `10.0.0.0/8" accept` },
		"rule note with newline":  func(d *Document) { *d = legacy(); d.Firewall[0].Note = "a\nb" },
		"rule and segmentation":   func(d *Document) { d.Firewall = legacy().Firewall },
		"policy action":           func(d *Document) { d.Segmentation.Policies[0].Action = "accept" },
		"group of a stranger":     func(d *Document) { d.Segmentation.Groups[0].Nodes = []string{"ghost"} },
		"zone member injection":   func(d *Document) { d.Segmentation.Zones[0].Members = []string{"10.77.1.0/24\nrole: server"} },
		"schedule injection":      func(d *Document) { d.Segmentation.Policies[0].Schedule.Days = []string{"sun\nrole: server"} },
		"time zone injection":     func(d *Document) { d.Segmentation.TimeZone = "../../etc/passwd" },
	}
	for name, edit := range cases {
		d := FromSnapshot(snapshot())
		edit(&d)
		if _, err := ToSnapshot(d); err == nil || !strings.Contains(err.Error(), "nothing was imported") {
			t.Errorf("%s: ToSnapshot = %v", name, err)
		}
	}
}

// dualSnapshot is snapshot() with IPv6 (M8): the prefix, the NAS's IPv6
// address and containers, an IPv6 endpoint on vps-1, an IPv6-only laptop
// and an IPv6 record.
func dualSnapshot() store.Snapshot {
	s := snapshot()
	s.Settings[store.SettingMeshCIDR6] = testfix.MeshCIDR6
	s.Settings[store.SettingUpstreamDNS] = "2606:4700:4700::1111,9.9.9.9"
	s.Nodes[0].MeshIP6, s.Nodes[0].ContainerSubnet6 = testfix.NAS.MeshIP6, testfix.ContainerSubnet6
	s.Nodes[1].MeshIP6, s.Nodes[1].Endpoint6 = testfix.VPS1.MeshIP6, "[2001:db8::10]:51820"
	s.Nodes = append(s.Nodes, store.SnapNode{ID: 5, Name: "laptop", Role: "spoke", MeshIP6: testfix.Laptop.MeshIP6, Family: "ipv6",
		ListenPort: 51820, Keepalive: 25, MTU: 1420})
	s.DNS = append(s.DNS, store.DNSRecord{ID: 2, Name: "printer", IP6: "fd10:77::1:200"})
	return s
}

func TestDualStackExport(t *testing.T) {
	in := dualSnapshot()
	var b bytes.Buffer
	if err := Write(&b, FromSnapshot(in)); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"mesh_cidr6: " + testfix.MeshCIDR6, "mesh_ip6: " + testfix.NAS.MeshIP6, "container_subnet6: " + testfix.ContainerSubnet6,
		"[2001:db8::10]:51820", "ip_family: ipv6", "ip6: fd10:77::1:200"} {
		if !strings.Contains(out, want) {
			t.Errorf("export lacks %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "mesh_ip:") != 2 {
		t.Errorf("the IPv6-only laptop must have no mesh_ip:\n%s", out)
	}
	d, err := Read(&b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ToSnapshot(d)
	if err != nil {
		t.Fatal(err)
	}
	want := []store.SnapNode{in.Nodes[1], in.Nodes[0], in.Nodes[2]} // by id
	if !reflect.DeepEqual(got.Nodes, want) || got.DNS[1].IP6 != "fd10:77::1:200" || got.DNS[1].IP != "" ||
		got.Settings[store.SettingMeshCIDR6] != testfix.MeshCIDR6 || got.Settings[store.SettingUpstreamDNS] != in.Settings[store.SettingUpstreamDNS] {
		t.Fatalf("round trip:\n%+v\nwant\n%+v", got.Nodes, want)
	}

	for name, edit := range map[string]func(*Document){
		"bad mesh ip6":         func(d *Document) { d.Nodes[1].MeshIP6 = "fd10:77::1:1:2:3:4:5:6" },
		"mesh ip6 is IPv4":     func(d *Document) { d.Nodes[1].MeshIP6 = "10.77.1.1" },
		"mesh ip6 injection":   func(d *Document) { d.Nodes[1].MeshIP6 = "fd10:77::1:1\nrole: server" },
		"no address":           func(d *Document) { d.Nodes[2].MeshIP6 = "" },
		"unknown family":       func(d *Document) { d.Nodes[2].Family = "both" },
		"endpoint6 is IPv4":    func(d *Document) { d.Nodes[0].Endpoint6 = testfix.VPS1.Endpoint },
		"container6 injection": func(d *Document) { d.Nodes[1].ContainerSubnet6 = testfix.ContainerSubnet6 + "\nrole: server" },
		"container6 is IPv4":   func(d *Document) { d.Nodes[1].ContainerSubnet6 = testfix.ContainerSubnet },
		"record ip6 is IPv4":   func(d *Document) { d.DNS[1].IP6 = "10.77.1.200" },
		"record with node ip6": func(d *Document) { d.DNS[0].IP6 = "fd10:77::1:9" },
		"mesh cidr6 host bits": func(d *Document) { d.Settings[store.SettingMeshCIDR6] = "fd10:77::1/64" },
		"upstream injection":   func(d *Document) { d.Settings[store.SettingUpstreamDNS] = "2606:4700:4700::1111;id" },
	} {
		d := FromSnapshot(dualSnapshot())
		edit(&d)
		if _, err := ToSnapshot(d); err == nil || !strings.Contains(err.Error(), "nothing was imported") {
			t.Errorf("%s: ToSnapshot = %v", name, err)
		}
	}
}

// An export made before the mesh firewall imports its custom rules as
// policies, in their order, above allow-all.
func TestToSnapshotReadsAnOlderExport(t *testing.T) {
	got, err := ToSnapshot(legacy())
	if err != nil {
		t.Fatal(err)
	}
	if got.Segmentation != nil || len(got.Firewall) != 2 || got.Firewall[0].NodeID != 1 || got.Firewall[1].ID != 2 {
		t.Fatalf("ToSnapshot = %+v", got)
	}
	seg := got.EffectiveSegmentation()
	if len(seg.Policies) != 3 || seg.Policies[0].Name != "firewall-1" || seg.Policies[0].To[0] != "node:vps-1" ||
		seg.Policies[1].To[0] != "role:server" || seg.Policies[2].Name != "allow-all" {
		t.Fatalf("restored as %+v", seg.Policies)
	}
}

// `meshctl firewall show --yaml` and `apply`: the document on its own.
func TestSegmentationFile(t *testing.T) {
	want := *snapshot().Segmentation
	var b bytes.Buffer
	if err := WriteSegmentation(&b, want); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(b.String(), "# meshctl mesh firewall") || !strings.Contains(b.String(), "segmentation:\n  groups:") {
		t.Fatalf("file:\n%s", b.String())
	}
	got, err := ReadSegmentation(&b)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("read back %+v, %v", got, err)
	}
	for name, in := range map[string]string{
		"unknown field": "segmentation:\n  rules: []\n",
		"bad action":    "segmentation:\n  policies:\n    - name: x\n      action: accept\n",
		"not yaml":      "segmentation: [\n",
	} {
		if _, err := ReadSegmentation(strings.NewReader(in)); err == nil {
			t.Errorf("%s: ReadSegmentation accepted it", name)
		}
	}
	if _, err := ReadSegmentation(strings.NewReader("segmentation:\n  policies:\n    - name: x\n      action: accept\n")); err == nil ||
		!strings.Contains(err.Error(), "nothing was changed") {
		t.Errorf("a bad file's error = %v", err)
	}
}
