package wg

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"
)

// The userspace backend talks to wireguard-go in WireGuard's cross-platform
// configuration protocol (https://www.wireguard.com/xplatform/): key=value
// lines, keys in hex. These two functions are its pure half, so the text is
// pinned by a golden test without a device.

// UAPISet renders c as a "set" operation. Nil fields are left out, so the
// device keeps them; peers are written in c's order, each one after its
// public_key line.
func UAPISet(c Change) string {
	var b strings.Builder
	line := func(k, v string) {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v)
		b.WriteByte('\n')
	}
	if c.PrivateKey != nil {
		line("private_key", hex.EncodeToString(c.PrivateKey[:]))
	}
	if c.ListenPort != nil {
		line("listen_port", strconv.Itoa(*c.ListenPort))
	}
	if c.FwMark != nil {
		line("fwmark", strconv.Itoa(*c.FwMark))
	}
	for _, p := range c.Peers {
		line("public_key", hex.EncodeToString(p.PublicKey[:]))
		if p.Remove {
			line("remove", "true")
			continue
		}
		if p.Endpoint != nil {
			line("endpoint", p.Endpoint.String())
		}
		if p.Keepalive != nil {
			line("persistent_keepalive_interval", strconv.Itoa(int(*p.Keepalive/time.Second)))
		}
		if p.ReplaceAllowedIPs {
			line("replace_allowed_ips", "true")
		}
		for _, a := range p.AllowedIPs {
			line("allowed_ip", a.String())
		}
	}
	return b.String()
}

// errHexKey never quotes the value: it may be a private key.
var errHexKey = errors.New("not a 64-digit hex key")

func hexKey(s string) (Key, error) {
	var k Key
	if len(s) != 2*len(k) {
		return k, errHexKey
	}
	if _, err := hex.Decode(k[:], []byte(s)); err != nil {
		return k, errHexKey
	}
	return k, nil
}

func uapiInt(s string, limit uint64) (int, error) {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n > limit {
		return 0, fmt.Errorf("%q is not a number from 0 to %d", s, limit)
	}
	return int(n), nil
}

// ParseUAPI reads the answer to a "get" operation: the device, as
// Host.Device returns it, and each peer's statistics. Peers are sorted by
// public key. Keys the agent does not use (preshared_key, protocol_version,
// ...) are skipped. A malformed value is an error naming its line and key;
// the value of a key line is never quoted.
func ParseUAPI(s string) (Device, []PeerStats, error) {
	var d Device
	var stats []PeerStats
	type handshake struct{ sec, nsec int64 }
	var times []handshake
	for i, ln := range strings.Split(s, "\n") {
		if ln == "" {
			continue
		}
		k, v, ok := strings.Cut(ln, "=")
		if !ok {
			return Device{}, nil, fmt.Errorf("read the userspace device: line %d has no '='", i+1)
		}
		n := len(d.Peers) - 1
		var err error
		switch {
		case k == "public_key":
			var pk Key
			if pk, err = hexKey(v); err == nil {
				d.Peers = append(d.Peers, Peer{PublicKey: pk})
				stats = append(stats, PeerStats{PublicKey: pk})
				times = append(times, handshake{})
			}
		case n < 0:
			switch k {
			case "private_key":
				d.PrivateKey, err = hexKey(v)
			case "listen_port":
				d.ListenPort, err = uapiInt(v, 0xffff)
			case "fwmark":
				d.FwMark, err = uapiInt(v, 0xffffffff)
			}
		default:
			switch k {
			case "endpoint":
				var ep netip.AddrPort
				ep, err = netip.ParseAddrPort(v)
				d.Peers[n].Endpoint, stats[n].Endpoint = ep, ep
			case "persistent_keepalive_interval":
				var ka int
				ka, err = uapiInt(v, 0xffff)
				d.Peers[n].Keepalive = time.Duration(ka) * time.Second
			case "allowed_ip":
				var p netip.Prefix
				if p, err = netip.ParsePrefix(v); err == nil {
					d.Peers[n].AllowedIPs = append(d.Peers[n].AllowedIPs, p)
				}
			case "last_handshake_time_sec":
				times[n].sec, err = strconv.ParseInt(v, 10, 64)
			case "last_handshake_time_nsec":
				times[n].nsec, err = strconv.ParseInt(v, 10, 64)
			case "rx_bytes":
				stats[n].RxBytes, err = strconv.ParseInt(v, 10, 64)
			case "tx_bytes":
				stats[n].TxBytes, err = strconv.ParseInt(v, 10, 64)
			}
		}
		if err != nil {
			return Device{}, nil, fmt.Errorf("read the userspace device: line %d, %s: %w", i+1, k, err)
		}
	}
	for i := range d.Peers {
		SortPrefixes(d.Peers[i].AllowedIPs)
		if t := times[i]; t.sec != 0 || t.nsec != 0 {
			stats[i].LastHandshake = time.Unix(t.sec, t.nsec)
		}
	}
	slices.SortFunc(d.Peers, func(a, b Peer) int { return compareKeys(a.PublicKey, b.PublicKey) })
	slices.SortFunc(stats, func(a, b PeerStats) int { return compareKeys(a.PublicKey, b.PublicKey) })
	return d, stats, nil
}
