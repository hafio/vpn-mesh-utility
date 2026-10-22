package server

import (
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// serverZone names the server's time zone, the default the mesh firewall's
// schedules are read in (amendment A43): loc's name when TZ gave it one,
// else the zone /etc/localtime links to (readlink), else UTC. Go calls the
// zone of /etc/localtime "Local", which names no zone an agent could load.
func serverZone(loc *time.Location, readlink func(string) (string, error)) string {
	name := loc.String()
	if name == "Local" {
		if target, err := readlink("/etc/localtime"); err == nil {
			if _, zone, ok := strings.Cut(target, "zoneinfo/"); ok {
				name = zone
			}
		}
	}
	if _, err := validate.TimeZone("the server's time zone", name); err != nil {
		return "UTC"
	}
	return name
}
