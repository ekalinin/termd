package termbg

import (
	"io"
	"math"
	"regexp"
	"strconv"
	"time"
)

// Terminal is a terminal connection that can wait for input.
type Terminal interface {
	io.Writer
	// ReadTimeout reads available input, waiting at most d. It returns
	// 0 and a nil error when nothing arrived in time.
	ReadTimeout(p []byte, d time.Duration) (int, error)
}

// query asks for the background color (OSC 11) and then for the device
// attributes (DA1). Every terminal answers DA1, so its reply marks the end of
// the answers even when OSC 11 is not supported.
const query = "\x1b]11;?\x1b\\\x1b[c"

var (
	oscRE = regexp.MustCompile(`\x1b\]11;rgba?:([0-9a-fA-F]{1,4})/([0-9a-fA-F]{1,4})/([0-9a-fA-F]{1,4})`)
	da1RE = regexp.MustCompile(`\x1b\[\?[0-9;]*c`)
)

// Detect queries t and reports whether its background is light. ok is false
// when the terminal did not report a background color within timeout.
func Detect(t Terminal, timeout time.Duration) (light, ok bool) {
	if _, err := io.WriteString(t, query); err != nil {
		return false, false
	}
	deadline := time.Now().Add(timeout)
	var buf []byte
	chunk := make([]byte, 256)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		n, err := t.ReadTimeout(chunk, remaining)
		if err != nil || n == 0 {
			break
		}
		buf = append(buf, chunk[:n]...)
		if da1RE.Match(buf) {
			break
		}
	}
	m := oscRE.FindSubmatch(buf)
	if m == nil {
		return false, false
	}
	return Luminance(channel(m[1]), channel(m[2]), channel(m[3])) >= 0.5, true
}

// channel converts a 1-4 digit hex color component to 0..1.
func channel(hex []byte) float64 {
	v, err := strconv.ParseUint(string(hex), 16, 16)
	if err != nil {
		return 0
	}
	return float64(v) / float64(uint64(1)<<(4*len(hex))-1)
}

// Luminance returns the relative luminance of an sRGB color with components
// in 0..1.
func Luminance(r, g, b float64) float64 {
	lin := func(c float64) float64 {
		if c <= 0.04045 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}
