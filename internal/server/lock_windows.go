//go:build windows

package server

import "errors"

// Lock is Linux-only: the server and import run where the data dir lives, in
// the server's container.
func Lock(string) (func(), error) {
	return nil, errors.New("the meshctl server and import run on Linux, in the server's container")
}
