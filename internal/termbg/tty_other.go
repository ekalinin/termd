//go:build !(darwin || linux || freebsd || netbsd || openbsd || dragonfly)

package termbg

import "time"

// Light is not supported on this platform; callers fall back to the dark
// theme.
func Light(timeout time.Duration) (light, ok bool) {
	return false, false
}
