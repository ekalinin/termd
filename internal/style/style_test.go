package style

import (
	"strings"
	"testing"
)

func TestPlainModeEmitsNothing(t *testing.T) {
	o := Options{}
	for _, s := range []Style{
		{Bold: true}, {Italic: true}, {Strike: true}, {ANSI: 36}, {FG: RGB(1, 2, 3)},
	} {
		if got := o.Format("text", s); got != "text" {
			t.Errorf("Format(%+v) = %q, want plain text", s, got)
		}
	}
	if o.LinkOpen("https://example.com") != "" || o.LinkClose() != "" {
		t.Error("link sequences emitted with hyperlinks disabled")
	}
}

func TestStyledMode(t *testing.T) {
	o := Options{Styled: true}
	tests := []struct {
		s    Style
		want string
	}{
		{Style{Bold: true}, "\x1b[1m"},
		{Style{Italic: true}, "\x1b[3m"},
		{Style{Strike: true}, "\x1b[9m"},
		{Style{ANSI: 36}, "\x1b[36m"},
		{Style{Bold: true, Underline: true}, "\x1b[1;4m"},
	}
	for _, tt := range tests {
		got := o.Format("x", tt.s)
		if got != tt.want+"x"+Reset {
			t.Errorf("Format(%+v) = %q, want %q", tt.s, got, tt.want+"x"+Reset)
		}
	}
	if got := o.Format("x", Style{}); got != "x" {
		t.Errorf("zero style emitted %q", got)
	}
}

func TestColorDepth(t *testing.T) {
	fg := Style{FG: RGB(255, 123, 114)}
	if got := (Options{Styled: true, Depth: TrueColor}).SGR(fg); got != "\x1b[38;2;255;123;114m" {
		t.Errorf("truecolor SGR = %q", got)
	}
	got := (Options{Styled: true, Depth: Color256}).SGR(fg)
	if !strings.HasPrefix(got, "\x1b[38;5;") || strings.Contains(got, "38;2;") {
		t.Errorf("256-color SGR = %q", got)
	}
}

func TestTo256(t *testing.T) {
	tests := []struct {
		r, g, b uint8
		want    int
	}{
		{0, 0, 0, 16},
		{255, 255, 255, 231},
		{255, 0, 0, 196},
		{128, 128, 128, 244},
	}
	for _, tt := range tests {
		if got := To256(tt.r, tt.g, tt.b); got != tt.want {
			t.Errorf("To256(%d,%d,%d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.want)
		}
	}
}

func TestLayer(t *testing.T) {
	tests := []struct {
		name      string
		base, top Style
		want      Style
	}{
		{"attributes are combined", Style{Bold: true, Faint: true}, Style{Italic: true, Underline: true, Strike: true},
			Style{Bold: true, Faint: true, Italic: true, Underline: true, Strike: true}},
		{"basic color replaces 24-bit color", Style{Bold: true, FG: RGB(1, 2, 3)}, Style{ANSI: 36},
			Style{Bold: true, ANSI: 36}},
		{"24-bit color replaces basic color", Style{Underline: true, ANSI: 34}, Style{FG: RGB(1, 2, 3)},
			Style{Underline: true, FG: RGB(1, 2, 3)}},
		{"top without color keeps base color", Style{ANSI: 34}, Style{Bold: true},
			Style{Bold: true, ANSI: 34}},
		{"zero top returns base", Style{Bold: true, FG: RGB(1, 2, 3)}, Style{},
			Style{Bold: true, FG: RGB(1, 2, 3)}},
	}
	for _, tt := range tests {
		if got := Layer(tt.base, tt.top); got != tt.want {
			t.Errorf("%s: Layer(%+v, %+v) = %+v, want %+v", tt.name, tt.base, tt.top, got, tt.want)
		}
	}
}

func TestLinkSequences(t *testing.T) {
	o := Options{Hyperlinks: true}
	if got := o.LinkOpen("https://example.com"); got != "\x1b]8;;https://example.com\x1b\\" {
		t.Errorf("LinkOpen = %q", got)
	}
	if got := o.LinkClose(); got != "\x1b]8;;\x1b\\" {
		t.Errorf("LinkClose = %q", got)
	}
}
