package proto_test

import (
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func validNames() proto.NamesRequest {
	return proto.NamesRequest{Names: []proto.ContainerName{
		{Name: "Web", Kind: proto.NameService, IP: "10.77.128.129"},
		{Name: "web.example.home.arpa.", Kind: proto.NameHostname, IP: "10.77.128.129"},
		{Name: "web", Kind: proto.NameService, IP: "10.77.128.130"},
	}}
}

// A name is folded to lower case and loses its trailing dot.
func TestNamesRequestNormalises(t *testing.T) {
	r := validNames()
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	if r.Names[0].Name != "web" || r.Names[1].Name != "web.example.home.arpa" {
		t.Fatalf("names not normalised: %+v", r.Names)
	}
	if err := (&proto.NamesRequest{}).Validate(); err != nil {
		t.Fatalf("an empty list clears the names: %v", err)
	}
}

// Every bad entry is named; a name that would add a line to a log or a
// label to the zone is refused.
func TestNamesRequestRejects(t *testing.T) {
	for name, c := range map[string]struct {
		edit func(*proto.NamesRequest)
		want string
	}{
		"newline in a name":  {func(r *proto.NamesRequest) { r.Names[0].Name = "web\nx" }, "names[0].name"},
		"underscore":         {func(r *proto.NamesRequest) { r.Names[0].Name = "my_web" }, "names[0].name"},
		"empty name":         {func(r *proto.NamesRequest) { r.Names[2].Name = "" }, "names[2].name"},
		"unknown kind":       {func(r *proto.NamesRequest) { r.Names[1].Kind = "label" }, "names[1].kind"},
		"address not IPv4":   {func(r *proto.NamesRequest) { r.Names[2].IP = "fd00::1" }, "names[2].ip"},
		"listed twice":       {func(r *proto.NamesRequest) { r.Names[2].IP = "10.77.128.129" }, "listed twice"},
		"twice, other kinds": {func(r *proto.NamesRequest) { r.Names[2].IP, r.Names[2].Kind = "10.77.128.129", proto.NameAlias }, "listed twice"},
		"too many": {func(r *proto.NamesRequest) {
			r.Names = make([]proto.ContainerName, proto.MaxNames+1)
		}, "at most 512"},
	} {
		t.Run(name, func(t *testing.T) {
			r := validNames()
			c.edit(&r)
			if err := r.Validate(); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Validate = %v, want an error naming %s", err, c.want)
			}
		})
	}
}
