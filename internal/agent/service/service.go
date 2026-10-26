// Package service runs the agent under the Windows service manager. The
// binary installs no service (amendment A51): install.sh installs the
// systemd unit meshctl-agent@.service, install.ps1 the Windows service
// meshctl-agent-<mesh>, each running `meshctl agent run --mesh <mesh>`.
package service

// Name is the Windows service of the agent of mesh, as install.ps1 names
// it.
func Name(mesh string) string { return "meshctl-agent-" + mesh }
