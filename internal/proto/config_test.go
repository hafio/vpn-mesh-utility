package proto_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// goldenConfig is the config whose canonical encoding and hash are pinned
// below. Changing NodeConfig's field names, order or omitempty rules changes
// the hash of every deployed config, which makes every agent re-apply; this
// test makes that a deliberate decision.
func goldenConfig() proto.NodeConfig {
	port := 51820
	return proto.NodeConfig{
		Version: 7,
		Node:    proto.Self{ID: 1, Name: "vps-1", Role: proto.RoleServer},
		Interface: proto.Interface{
			Address:    "10.77.0.1/32",
			ListenPort: &port,
			MTU:        1420,
		},
		Peers: []proto.Peer{{
			Name:       "vps-2",
			PublicKey:  testfix.WGKey(2),
			Endpoint:   "203.0.113.11:51820",
			AllowedIPs: []string{"10.77.0.2/32"},
		}},
		Firewall: &proto.Firewall{
			Backend: "firewalld",
			Rules:   []proto.FWRule{{Zone: "public", Proto: "udp", Port: 51820, Note: "wireguard"}},
		},
	}
}

const (
	goldenCanonical = `{"version":0,"node":{"id":1,"name":"vps-1","role":"server"},"interface":{"address":"10.77.0.1/32","listen_port":51820,"mtu":1420},"peers":[{"name":"vps-2","public_key":"AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI=","endpoint":"203.0.113.11:51820","allowed_ips":["10.77.0.2/32"]}],"firewall":{"backend":"firewalld","rules":[{"zone":"public","proto":"udp","port":51820,"note":"wireguard"}]}}`
	// printf '%s' "$goldenCanonical" | sha256sum
	goldenHash = "371e8e9be3873930c400381117f7c78275a2882f0e736402ba770b03af655a29"
)

func TestHashIsPinned(t *testing.T) {
	c := goldenConfig()
	c.Version = 0
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != goldenCanonical {
		t.Fatalf("canonical encoding changed:\n got %s\nwant %s", b, goldenCanonical)
	}
	got, err := goldenConfig().Hash()
	if err != nil {
		t.Fatal(err)
	}
	if got != goldenHash {
		t.Fatalf("Hash = %s, want %s", got, goldenHash)
	}
}

func TestHashIgnoresVersionOnly(t *testing.T) {
	a := goldenConfig()
	b := goldenConfig()
	b.Version = 99
	ha, _ := a.Hash()
	hb, _ := b.Hash()
	if ha != hb {
		t.Fatal("Hash must not depend on Version")
	}
	b.Interface.MTU = 1380
	hc, _ := b.Hash()
	if hc == ha {
		t.Fatal("Hash must change when the desired state changes")
	}
}

func withDomains(c proto.NodeConfig, ds ...string) proto.NodeConfig {
	c.DNS = &proto.DNS{Zone: proto.DefaultZone, Split: true, Domains: ds}
	return c
}

func TestValidateAcceptsFixtures(t *testing.T) {
	// A host-mode server whose containers are on the mesh: its firewall
	// forwards to their bridge (amendment A37).
	hosting := testfix.ServerConfig()
	hosting.Containers = testfix.Containers(proto.ModeHost)
	hosting.Firewall.Bridge, hosting.Firewall.Relay = proto.ContainerBridge, true
	hosting.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}
	// A server with the mesh firewall: its mesh zone accepts the mesh, the
	// policy decides (amendment A43).
	firewalled := testfix.ServerConfig()
	firewalled.Firewall.Rules[0] = proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Source: testfix.MeshCIDR, Note: "the policy decides"}
	firewalled.Policy = testfix.Policy()
	for name, c := range map[string]proto.NodeConfig{
		"golden":                     goldenConfig(),
		"testfix server":             testfix.ServerConfig(),
		"testfix spoke":              testfix.SpokeConfig(""),
		"testfix spoke with an exit": testfix.SpokeConfig(testfix.VPS2.Name),
		"container-mode spoke":       testfix.ContainerSpokeConfig(),
		"server with containers":     hosting,
		"extra domains":              withDomains(testfix.SpokeConfig(""), "example.lan", "home.arpa"),
		"server with a policy":       firewalled,
	} {
		c := c
		if err := c.Validate(); err != nil {
			t.Errorf("%s: Validate = %v", name, err)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*proto.NodeConfig)
		want   string
	}{
		{"negative version", func(c *proto.NodeConfig) { c.Version = -1 }, "version"},
		{"bad node name", func(c *proto.NodeConfig) { c.Node.Name = "VPS 1" }, "node.name"},
		{"unknown role", func(c *proto.NodeConfig) { c.Node.Role = "admin" }, "node.role"},
		{"address not a /32", func(c *proto.NodeConfig) { c.Interface.Address = "10.77.0.0/24" }, "must be a /32"},
		{"address garbage", func(c *proto.NodeConfig) { c.Interface.Address = "10.77.0.1;reboot" }, "interface.address"},
		{"mtu too small", func(c *proto.NodeConfig) { c.Interface.MTU = 576 }, "interface.mtu"},
		{"listen port zero", func(c *proto.NodeConfig) { p := 0; c.Interface.ListenPort = &p }, "interface.listen_port"},
		{"negative fwmark", func(c *proto.NodeConfig) { c.Interface.FwMark = -1 }, "interface.fwmark"},
		{"bad peer name", func(c *proto.NodeConfig) { c.Peers[0].Name = "vps 2" }, "peers[0].name"},
		{"duplicate peer name", func(c *proto.NodeConfig) { c.Peers[1].Name = c.Peers[0].Name }, "duplicate peer"},
		{"duplicate peer key", func(c *proto.NodeConfig) { c.Peers[1].PublicKey = c.Peers[0].PublicKey }, "same key"},
		{"prefix on two peers", func(c *proto.NodeConfig) { c.Peers[1].AllowedIPs = c.Peers[0].AllowedIPs }, "already routed"},
		{"empty allowed ips", func(c *proto.NodeConfig) { c.Peers[0].AllowedIPs = nil }, "must not be empty"},
		{"host bits in allowed ip", func(c *proto.NodeConfig) { c.Peers[0].AllowedIPs = []string{"10.77.0.2/24"} }, "host bits"},
		{"endpoint hostname", func(c *proto.NodeConfig) { c.Peers[0].Endpoint = "evil.example.com:51820" }, "endpoint"},
		{"bad key", func(c *proto.NodeConfig) { c.Peers[0].PublicKey = "not-a-key" }, "public_key"},
		{"keepalive too long", func(c *proto.NodeConfig) { c.Peers[0].Keepalive = 86400 }, "keepalive"},
		{"exit via stranger", func(c *proto.NodeConfig) { c.Exit = &proto.Exit{Via: "vps-9"} }, "not a peer"},
		{"exit via peer without default", func(c *proto.NodeConfig) { c.Exit = &proto.Exit{Via: c.Peers[0].Name} }, "does not carry 0.0.0.0/0"},
		{"exit choice", func(c *proto.NodeConfig) { c.Exit = &proto.Exit{Choices: []string{"vps-2;reboot"}} }, "exit.choices[0]"},
		{"unknown sysctl", func(c *proto.NodeConfig) {
			c.Sysctl = []proto.Sysctl{{Key: "kernel.core_pattern", Value: "|/bin/sh"}}
		}, "sysctl[0].key"},
		{"sysctl value", func(c *proto.NodeConfig) { c.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "2"}} }, "sysctl[0].value"},
		{"firewall backend", func(c *proto.NodeConfig) { c.Firewall.Backend = "iptables" }, "firewall.backend"},
		{"firewall proto", func(c *proto.NodeConfig) { c.Firewall.Rules[0].Proto = "sctp" }, "firewall.rules[0].proto"},
		{"firewall port", func(c *proto.NodeConfig) { c.Firewall.Rules[0].Port = 0 }, "firewall.rules[0].port"},
		{"firewall zone", func(c *proto.NodeConfig) { c.Firewall.Rules[0].Zone = "trusted" }, "zone"},
		{"firewall source", func(c *proto.NodeConfig) { c.Firewall.Rules[0].Source = "0.0.0.0/0 accept" }, "source"},
		{"firewall note newline", func(c *proto.NodeConfig) { c.Firewall.Rules[0].Note = "x\nrule accept" }, "note"},
		{"ping rule with a port", func(c *proto.NodeConfig) {
			c.Firewall.Rules[0] = proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoICMP, Port: 22, Source: testfix.MeshCIDR}
		}, "a ping rule has no port"},
		{"accept-all with a port", func(c *proto.NodeConfig) {
			c.Firewall.Rules[0] = proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Port: 22, Source: testfix.MeshCIDR}
		}, "an accept-all rule"},
		{"accept-all in the public zone", func(c *proto.NodeConfig) {
			c.Firewall.Rules[0] = proto.FWRule{Zone: proto.ZonePublic, Proto: proto.ProtoAll, Source: testfix.MeshCIDR}
		}, "an accept-all rule"},
		{"accept-all from anywhere", func(c *proto.NodeConfig) {
			c.Firewall.Rules[0] = proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll}
		}, "an accept-all rule"},
		{"policy on a spoke", func(c *proto.NodeConfig) {
			c.Node.Role, c.Policy = proto.RoleSpoke, testfix.Policy()
		}, "only a server enforces"},
		{"policy action", func(c *proto.NodeConfig) {
			c.Policy = testfix.Policy()
			c.Policy.Rules[3].Action = "maybe"
		}, "policy.rules[3].action"},
		{"dns zone", func(c *proto.NodeConfig) { c.DNS = &proto.DNS{Zone: "corp.local"} }, "dns.zone"},
		{"dns record ip", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, Records: []proto.Record{{Name: "nas", IP: "nas.local"}}}
		}, "dns.records[0].ip"},
		{"dns record name", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, Records: []proto.Record{{Name: "nas;reboot", IP: "10.77.1.1"}}}
		}, "dns.records[0].name"},
		{"dns upstream", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, Upstream: []string{"dns.google"}}
		}, "dns.upstream[0]"},
		{"dns forward_for", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, ForwardFor: []string{"10.77.3.0/33"}}
		}, "dns.forward_for[0]"},
		{"dns domain inside the zone", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, Domains: []string{"lab.mesh"}}
		}, "dns.domains"},
		{"dns domains unsorted", func(c *proto.NodeConfig) {
			c.DNS = &proto.DNS{Zone: proto.DefaultZone, Domains: []string{"home.arpa", "example.lan"}}
		}, "sorted"},
		{"firewall bridge", func(c *proto.NodeConfig) { c.Firewall.Bridge = "docker0" }, "firewall.bridge"},
		{"containers mode", func(c *proto.NodeConfig) { c.Containers = withMode("bridge") }, "containers.mode"},
		{"containers subnet garbage", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Containers.Subnet = "10.77.128.0/24;reboot"
		}, "containers.subnet"},
		{"containers subnet not a /24", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Containers.Subnet = "10.77.128.0/25"
		}, "must be a /24"},
		{"containers network without the subnet", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Containers.Network = "10.88.0.0/16"
		}, "does not contain the subnet"},
		{"containers range outside the subnet", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Containers.Range = "10.77.129.128/25"
		}, "is not a part of the subnet"},
		{"containers gateway in the range", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeContainer)
			c.Containers.Gateway = "10.77.128.130"
		}, "outside the range"},
		{"host mode gateway is not the bridge", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Containers.Gateway = "10.77.128.2"
		}, "the gateway is the bridge"},
		{"container mode gateway is the bridge", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeContainer)
			c.Containers.Gateway = c.Containers.Bridge
		}, "the gateway is the agent"},
		{"containers subnet holds the node", func(c *proto.NodeConfig) {
			c.Containers = testfix.Containers(proto.ModeHost)
			c.Interface.Address = "10.77.128.9/32"
		}, "holds the node's own mesh IP"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := testfix.ServerConfig()
			c.mutate(&cfg)
			err := cfg.Validate()
			if err == nil {
				t.Fatal("Validate accepted an invalid config")
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Validate error %q does not mention %q", err, c.want)
			}
		})
	}
}

func withMode(mode string) *proto.Containers {
	k := testfix.Containers(proto.ModeHost)
	k.Mode = mode
	return k
}

func TestValidateReportsEveryProblem(t *testing.T) {
	cfg := testfix.ServerConfig()
	cfg.Node.Name = "Bad Name"
	cfg.Interface.MTU = 9000
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "node.name") || !strings.Contains(err.Error(), "interface.mtu") {
		t.Fatalf("Validate must report all problems, got %v", err)
	}
}
