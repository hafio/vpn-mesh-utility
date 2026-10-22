package server

import (
	"errors"
	"testing"
	"time"
)

func TestServerZone(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	local := time.FixedZone("Local", 0)
	link := func(target string, err error) func(string) (string, error) {
		return func(path string) (string, error) {
			if path != "/etc/localtime" {
				t.Errorf("readlink %q", path)
			}
			return target, err
		}
	}
	for _, c := range []struct {
		name     string
		loc      *time.Location
		readlink func(string) (string, error)
		want     string
	}{
		{"TZ names it", berlin, link("", errors.New("not called")), "Europe/Berlin"},
		{"UTC", time.UTC, link("", errors.New("not called")), "UTC"},
		{"the link", local, link("../usr/share/zoneinfo/Asia/Singapore", nil), "Asia/Singapore"},
		{"no link", local, link("", errors.New("invalid argument")), "UTC"},
		{"a link elsewhere", local, link("/opt/zones/custom", nil), "UTC"},
		{"a link to no zone", local, link("/usr/share/zoneinfo/Mars/Olympus", nil), "UTC"},
		{"a zone with no name", time.FixedZone("", 3600), link("", nil), "UTC"},
	} {
		if got := serverZone(c.loc, c.readlink); got != c.want {
			t.Errorf("%s: serverZone = %q, want %q", c.name, got, c.want)
		}
	}
}
