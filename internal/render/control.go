package render

import (
	"bytes"
	"unicode/utf8"
)

// cleanSource returns a copy of a document in which control characters
// cannot reach the terminal. It runs before anything is parsed or measured.
// CRLF and a lone CR are line endings, as in CommonMark, and become LF.
func cleanSource(src []byte) []byte {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
	src = bytes.ReplaceAll(src, []byte("\r"), []byte("\n"))
	return bytes.Map(controlPicture, src)
}

// controlPicture returns what is shown for r: a C0 control character other
// than tab and line feed becomes its control picture (ESC is ␛), DEL becomes
// ␡, and a C1 control character becomes U+FFFD. bytes.Map and strings.Map
// also turn every byte that is not valid UTF-8 into U+FFFD.
func controlPicture(r rune) rune {
	switch {
	case r == '\t' || r == '\n':
		return r
	case r < 0x20:
		return 0x2400 + r
	case r == 0x7f:
		return 0x2421
	case r >= 0x80 && r <= 0x9f:
		return utf8.RuneError
	}
	return r
}
