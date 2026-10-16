//go:build !windows

package service

import (
	"context"
	"testing"
)

// Off Windows no process is a Windows service, and running as one fails.
func TestNoWindowsServiceHere(t *testing.T) {
	if is, err := IsWindowsService(); is || err != nil {
		t.Fatalf("IsWindowsService = %v, %v", is, err)
	}
	if err := RunWindowsService(func(context.Context) error { return nil }); err == nil {
		t.Fatal("RunWindowsService succeeded off Windows")
	}
}
