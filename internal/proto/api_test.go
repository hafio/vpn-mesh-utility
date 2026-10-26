package proto_test

import (
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func validJoin() proto.JoinRequest {
	return proto.JoinRequest{
		Token:             testfix.Token,
		Hostname:          "VPS-2.example.com",
		OS:                "linux",
		Arch:              "amd64",
		WGBackend:         "kernel",
		WGPublicKey:       testfix.WGKey(2),
		IdentityPublicKey: testfix.IdentityKeyB64(2),
		Endpoint:          "203.0.113.11:51820",
		AgentVersion:      "v1.0.0",
		MachineID:         testfix.MachineID,
	}
}

func TestJoinRequestValidateNormalises(t *testing.T) {
	r := validJoin()
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if r.Hostname != "vps-2.example.com" {
		t.Fatalf("hostname not normalised: %q", r.Hostname)
	}
	r.PublicIP = "198.51.100.7"
	if err := r.Validate(); err != nil || r.PublicIP != "198.51.100.7" {
		t.Fatalf("public_ip: %v", err)
	}
	r.Endpoint6, r.PublicIP6 = "[2001:DB8::11]:51820", "2001:DB8::0011"
	if err := r.Validate(); err != nil || r.Endpoint6 != "[2001:db8::11]:51820" || r.PublicIP6 != "2001:db8::11" {
		t.Fatalf("the IPv6 side: %v, %q %q", err, r.Endpoint6, r.PublicIP6)
	}
}

func TestETagQuotesTheHash(t *testing.T) {
	if got := proto.ETag("abc"); got != `"abc"` {
		t.Fatalf("ETag = %s", got)
	}
}

func TestJoinRequestValidateRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*proto.JoinRequest)
		want   string
	}{
		{"token", func(r *proto.JoinRequest) { r.Token = "short" }, "token"},
		{"name injection", func(r *proto.JoinRequest) { r.Name = "nas\nprivileged: true" }, "name"},
		{"hostname injection", func(r *proto.JoinRequest) { r.Hostname = "$(reboot)" }, "hostname"},
		{"os", func(r *proto.JoinRequest) { r.OS = "darwin" }, "os"},
		{"arch", func(r *proto.JoinRequest) { r.Arch = "386" }, "arch"},
		{"backend", func(r *proto.JoinRequest) { r.WGBackend = "boringtun" }, "wg_backend"},
		{"wg key", func(r *proto.JoinRequest) { r.WGPublicKey = "x" }, "wg_public_key"},
		{"identity key", func(r *proto.JoinRequest) { r.IdentityPublicKey = "" }, "identity_public_key"},
		{"endpoint", func(r *proto.JoinRequest) { r.Endpoint = "vps2:51820" }, "endpoint"},
		{"public ip", func(r *proto.JoinRequest) { r.PublicIP = "203.0.113.11:51820" }, "public_ip"},
		{"endpoint6 is IPv4", func(r *proto.JoinRequest) { r.Endpoint6 = "203.0.113.11:51820" }, "endpoint6"},
		{"endpoint6 injection", func(r *proto.JoinRequest) { r.Endpoint6 = "[2001:db8::11]:51820\nx" }, "endpoint6"},
		{"public ip6 is IPv4", func(r *proto.JoinRequest) { r.PublicIP6 = "198.51.100.7" }, "public_ip6"},
		{"public ip6 with a zone", func(r *proto.JoinRequest) { r.PublicIP6 = "fe80::1%eth0" }, "public_ip6"},
		{"version", func(r *proto.JoinRequest) { r.AgentVersion = "1.0;rm" }, "agent_version"},
		{"no machine id", func(r *proto.JoinRequest) { r.MachineID = "" }, "machine_id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := validJoin()
			c.mutate(&r)
			err := r.Validate()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Validate = %v, want an error naming %q", err, c.want)
			}
		})
	}
}

func TestJoinRequestErrorNeverEchoesTheToken(t *testing.T) {
	r := validJoin()
	r.Token = "this-is-a-secret-token-value-but-malformed!!"
	err := r.Validate()
	if err == nil || strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("token must be masked in the error, got %v", err)
	}
}

func validStatus() proto.StatusReport {
	return proto.StatusReport{
		AppliedVersion: 3,
		AppliedHash:    strings.Repeat("a", 64),
		AgentVersion:   "v1.0.0",
		WGBackend:      "kernel",
		Peers:          []proto.PeerStatus{{Name: "vps-2", LastHandshake: testfix.Epoch.Unix(), RxBytes: 10, TxBytes: 20, Endpoint: "203.0.113.11:51820"}},
		Degraded:       []string{"dns"},
		DNS:            proto.DNSUnavailable,
	}
}

func TestStatusReportValidate(t *testing.T) {
	r := validStatus()
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	r.Families = []string{proto.FamilyIPv6, proto.FamilyIPv4}
	if err := r.Validate(); err != nil {
		t.Fatalf("Validate with both families: %v", err)
	}
	cases := []struct {
		name   string
		mutate func(*proto.StatusReport)
		want   string
	}{
		{"negative version", func(r *proto.StatusReport) { r.AppliedVersion = -1 }, "applied_version"},
		{"hash", func(r *proto.StatusReport) { r.AppliedHash = "XYZ" }, "applied_hash"},
		{"peer name", func(r *proto.StatusReport) { r.Peers[0].Name = "../x" }, "peers[0].name"},
		{"peer counters", func(r *proto.StatusReport) { r.Peers[0].RxBytes = -5 }, "counters"},
		{"peer endpoint", func(r *proto.StatusReport) { r.Peers[0].Endpoint = "nowhere" }, "endpoint"},
		{"too many peers", func(r *proto.StatusReport) { r.Peers = make([]proto.PeerStatus, 2000) }, "at most"},
		{"error with newline", func(r *proto.StatusReport) { r.Errors = []string{"a\nb"} }, "errors[0]"},
		{"too many errors", func(r *proto.StatusReport) { r.Errors = make([]string, 64) }, "at most"},
		{"unknown degraded step", func(r *proto.StatusReport) { r.Degraded = []string{"peers"} }, "degraded[0]"},
		{"rollback reason", func(r *proto.StatusReport) { r.Rollback = &proto.Rollback{From: 4, To: 3, Reason: "x\x00"} }, "rollback.reason"},
		{"unknown dns state", func(r *proto.StatusReport) { r.DNS = "maybe" }, "dns"},
		{"unknown family", func(r *proto.StatusReport) { r.Families = []string{"ipv4", "ipx"} }, "families[1]"},
		{"family twice", func(r *proto.StatusReport) { r.Families = []string{"ipv6", "ipv6"} }, "listed twice"},
		{"family injection", func(r *proto.StatusReport) { r.Families = []string{"ipv4\nipv6"} }, "families[0]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := validStatus()
			c.mutate(&r)
			err := r.Validate()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Validate = %v, want an error naming %q", err, c.want)
			}
		})
	}
}

func TestPrefsRequestValidate(t *testing.T) {
	str := func(s string) *string { return &s }
	num := func(n int) *int { return &n }
	for _, ok := range []proto.PrefsRequest{
		{Exit: str("vps-2")}, {Exit: str("")}, {Keepalive: num(0)}, {Keepalive: num(25), MTU: num(1380)},
	} {
		if err := ok.Validate(); err != nil {
			t.Errorf("Validate(%+v) = %v", ok, err)
		}
	}
	cases := []struct {
		name string
		r    proto.PrefsRequest
		want string
	}{
		{"nothing to change", proto.PrefsRequest{}, "nothing to change"},
		{"exit with a shell metacharacter", proto.PrefsRequest{Exit: str("vps-2;reboot")}, "exit"},
		{"keepalive too long", proto.PrefsRequest{Keepalive: num(3601)}, "keepalive"},
		{"mtu too small", proto.PrefsRequest{MTU: num(1000)}, "mtu"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.r.Validate(); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Validate = %v, want an error naming %q", err, c.want)
			}
		})
	}
}
