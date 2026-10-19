//go:build windows && amd64 && wintun

package wintun

import _ "embed"

// dll is the driver fetchwintun put beside this file.
//
//go:embed wintun_amd64.dll
var dll []byte

func bundled() []byte { return dll }
