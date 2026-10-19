//go:build !(windows && amd64 && wintun)

package wintun

// bundled is nothing: only the Windows release build embeds the driver.
func bundled() []byte { return nil }
