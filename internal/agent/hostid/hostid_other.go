//go:build !windows

package hostid

// sources are where Linux keeps the machine id: systemd's, D-Bus's, and the
// firmware's product UUID where neither is set up.
func sources() []source {
	var out []source
	for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id", "/sys/class/dmi/id/product_uuid"} {
		out = append(out, source{name: p, read: func() (string, error) { return fromFile(p) }})
	}
	return out
}
