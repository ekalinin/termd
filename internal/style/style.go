package style

import (
	"strconv"
	"strings"
)

// Color is a 24-bit foreground color. The zero value means "no color".
type Color struct {
	R, G, B uint8
	Set     bool
}

// RGB returns a set color.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, Set: true}
}

// Style describes how a piece of text looks. The zero value is unstyled.
type Style struct {
	Bold      bool
	Italic    bool
	Underline bool
	Strike    bool
	Faint     bool
	// ANSI is a basic palette color code (30-37, 90-97); 0 means none.
	ANSI int
	// FG is a 24-bit color, converted to the 256-color palette when needed.
	FG Color
}

// IsZero reports whether the style has no attributes.
func (s Style) IsZero() bool {
	return s == Style{}
}

// Layer returns top applied over base: the attributes of both are combined,
// and the color of top, when it has one, replaces the color of base.
func Layer(base, top Style) Style {
	s := base
	s.Bold = s.Bold || top.Bold
	s.Italic = s.Italic || top.Italic
	s.Underline = s.Underline || top.Underline
	s.Strike = s.Strike || top.Strike
	s.Faint = s.Faint || top.Faint
	if top.ANSI != 0 || top.FG.Set {
		s.ANSI, s.FG = top.ANSI, top.FG
	}
	return s
}

// Depth is the color depth used for 24-bit colors.
type Depth int

const (
	// TrueColor emits 24-bit colors.
	TrueColor Depth = iota
	// Color256 emits the nearest color of the 256-color palette.
	Color256
)

// Options controls which escape sequences are emitted.
type Options struct {
	// Styled enables SGR sequences.
	Styled bool
	// Hyperlinks enables OSC 8 sequences.
	Hyperlinks bool
	Depth      Depth
}

// Reset ends an SGR sequence.
const Reset = "\x1b[0m"

// SGR returns the escape sequence that starts s, or "" when nothing is emitted.
func (o Options) SGR(s Style) string {
	if !o.Styled || s.IsZero() {
		return ""
	}
	var codes []string
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Faint {
		codes = append(codes, "2")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if s.Underline {
		codes = append(codes, "4")
	}
	if s.Strike {
		codes = append(codes, "9")
	}
	switch {
	case s.FG.Set && o.Depth == TrueColor:
		codes = append(codes, "38;2;"+itoa(int(s.FG.R))+";"+itoa(int(s.FG.G))+";"+itoa(int(s.FG.B)))
	case s.FG.Set:
		codes = append(codes, "38;5;"+itoa(To256(s.FG.R, s.FG.G, s.FG.B)))
	case s.ANSI != 0:
		codes = append(codes, itoa(s.ANSI))
	}
	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// Format wraps text in the SGR sequence of s.
func (o Options) Format(text string, s Style) string {
	sgr := o.SGR(s)
	if sgr == "" {
		return text
	}
	return sgr + text + Reset
}

// LinkOpen starts an OSC 8 hyperlink to url, or returns "" when disabled.
func (o Options) LinkOpen(url string) string {
	if !o.Hyperlinks || url == "" {
		return ""
	}
	return "\x1b]8;;" + url + "\x1b\\"
}

// LinkClose ends an OSC 8 hyperlink, or returns "" when disabled.
func (o Options) LinkClose() string {
	if !o.Hyperlinks {
		return ""
	}
	return "\x1b]8;;\x1b\\"
}

// cubeLevels are the channel values of the 6x6x6 color cube (indexes 16-231).
var cubeLevels = [6]int{0, 95, 135, 175, 215, 255}

// To256 returns the index of the 256-color palette entry closest to r, g, b,
// choosing between the color cube and the grayscale ramp (indexes 232-255).
func To256(r, g, b uint8) int {
	ri, gi, bi := nearestLevel(int(r)), nearestLevel(int(g)), nearestLevel(int(b))
	cube := 16 + 36*ri + 6*gi + bi
	cubeDist := dist(int(r), int(g), int(b), cubeLevels[ri], cubeLevels[gi], cubeLevels[bi])

	avg := (int(r) + int(g) + int(b)) / 3
	grayIdx := 0
	if avg > 8 {
		grayIdx = min((avg-8+5)/10, 23)
	}
	gv := 8 + 10*grayIdx
	grayDist := dist(int(r), int(g), int(b), gv, gv, gv)
	if grayDist < cubeDist {
		return 232 + grayIdx
	}
	return cube
}

func nearestLevel(v int) int {
	best, bestDiff := 0, 1<<30
	for i, l := range cubeLevels {
		d := v - l
		if d < 0 {
			d = -d
		}
		if d < bestDiff {
			best, bestDiff = i, d
		}
	}
	return best
}

func dist(r1, g1, b1, r2, g2, b2 int) int {
	dr, dg, db := r1-r2, g1-g2, b1-b2
	return dr*dr + dg*dg + db*db
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
