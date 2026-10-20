package proto_test

import (
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// The agent checks the server's mesh view before it caches or shows it.
func TestMeshViewValidate(t *testing.T) {
	v := testfix.MeshView()
	v.Names = testfix.MeshNames()
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		edit  func(*proto.MeshView)
		field string
	}{
		"negative version":   {func(v *proto.MeshView) { v.Version = -1 }, "version"},
		"name with newline":  {func(v *proto.MeshView) { v.Nodes[0].Name = "vps-1\nx" }, "nodes[0].name"},
		"unknown role":       {func(v *proto.MeshView) { v.Nodes[0].Role = "router" }, "nodes[0].role"},
		"mesh IP not IPv4":   {func(v *proto.MeshView) { v.Nodes[1].MeshIP = "fd00::1" }, "nodes[1].mesh_ip"},
		"hub injected":       {func(v *proto.MeshView) { v.Nodes[4].Hub = "vps-1;id" }, "nodes[4].hub"},
		"exit injected":      {func(v *proto.MeshView) { v.Nodes[4].Exit = "<b>" }, "nodes[4].exit"},
		"unknown health":     {func(v *proto.MeshView) { v.Nodes[0].Health = "green" }, "nodes[0].health"},
		"escape in state":    {func(v *proto.MeshView) { v.Nodes[0].State = "in sync\x1b[31m" }, "nodes[0].state"},
		"state too long":     {func(v *proto.MeshView) { v.Nodes[0].State = strings.Repeat("x", 201) }, "nodes[0].state"},
		"negative last seen": {func(v *proto.MeshView) { v.Nodes[1].LastSeen = -1 }, "nodes[1]"},
		"too many nodes":     {func(v *proto.MeshView) { v.Nodes = make([]proto.MeshNode, 1<<16+1) }, "at most 65536"},
		"name injected":      {func(v *proto.MeshView) { v.Names[0].Name = "db\nx" }, "names[0].name"},
		"name of no node":    {func(v *proto.MeshView) { v.Names[1].Node = "Laptop" }, "names[1].node"},
		"name not IPv4":      {func(v *proto.MeshView) { v.Names[2].IP = "web" }, "names[2].ip"},
		"too many names":     {func(v *proto.MeshView) { v.Names = make([]proto.MeshName, 1<<16+1) }, "names: at most"},
	} {
		t.Run(name, func(t *testing.T) {
			v := testfix.MeshView()
			v.Names = testfix.MeshNames()
			c.edit(&v)
			if err := v.Validate(); err == nil || !strings.Contains(err.Error(), c.field) {
				t.Fatalf("Validate = %v, want an error naming %s", err, c.field)
			}
		})
	}
}
