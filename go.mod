module github.com/hafio/vpn-mesh-utility

// `go` is the language floor; `toolchain` is the exact pin that local and CI
// both build with (scripts/dev.sh and dev.ps1 export GOTOOLCHAIN from it when
// it is unset). `go 1.27` alone is only a floor, so a runner with an older or
// newer patch release would otherwise build with a different stdlib. Bump both
// deliberately, with the gates re-run, and bump the Dockerfile's golang image
// in the same change (it builds with GOTOOLCHAIN=local).
go 1.27

toolchain go1.27.2

require (
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc
	github.com/godbus/dbus/v5 v5.2.3
	github.com/miekg/dns v1.1.73
	github.com/pquerna/otp v1.5.0
	github.com/vishvananda/netlink v1.3.1
	github.com/vishvananda/netns v0.0.5
	go.yaml.in/yaml/v3 v3.0.5
	golang.org/x/crypto v0.58.0
	golang.org/x/term v0.47.0
	golang.zx2c4.com/wireguard v0.0.0-20261006164505-2631ce99a06f
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20241231184526-a9ab2273dd10
	modernc.org/sqlite v1.60.1
)

require (
	github.com/BurntSushi/toml v1.4.1-0.20240526193622-a339e1f7089c // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20261010193011-f71dc92da45e // indirect
	golang.org/x/tools v0.52.0 // indirect
	golang.org/x/vuln v1.9.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	honnef.co/go/tools v0.8.1 // indirect
	modernc.org/libc v1.77.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
)

tool (
	golang.org/x/vuln/cmd/govulncheck
	honnef.co/go/tools/cmd/staticcheck
)
