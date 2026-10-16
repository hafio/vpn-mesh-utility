// Package server is a negative-control fixture: it crosses the boundary on
// purpose so archtest_test.go can prove the checker notices.
package server

import (
	_ "example.com/fake/internal/agent"
	_ "golang.zx2c4.com/wireguard/wgctrl"
)
