package site

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"

	"github.com/ekalinin/termd/internal/text"
)

// attrs is the text style set by SGR sequences.
type attrs struct {
	bold, faint, italic, underline, strike bool
	// ansi is a basic palette color code (30-37, 90-97); 0 means none.
	ansi int
	// rgb is a 24-bit color as "#rrggbb"; "" means none.
	rgb string
}

// segment is a run of text with one style and one link destination.
type segment struct {
	text  string
	attrs attrs
	link  string
}

// HTML converts termd's styled output to HTML for a <pre> block. It accepts
// only the escape sequences termd emits with 24-bit colors: SGR 0, 1, 2, 3,
// 4, 9, 30-37, 90-97 and 38;2;R;G;B, and OSC 8 hyperlinks terminated by
// ESC \. Any other escape sequence is an error. Every grapheme cluster two
// columns wide is wrapped in a two-cell box, so the alignment of the text
// does not depend on the width of its glyphs in the browser font.
func HTML(ansi string) (string, error) {
	segs, err := parse(ansi)
	if err != nil {
		return "", err
	}
	return emit(segs), nil
}

// parse splits s into segments of text with the style and link set by the
// escape sequences before them.
func parse(s string) ([]segment, error) {
	var segs []segment
	var cur attrs
	link := ""
	start := 0
	for i := 0; i < len(s); {
		if s[i] != '\x1b' {
			i++
			continue
		}
		if i > start {
			segs = append(segs, segment{text: s[start:i], attrs: cur, link: link})
		}
		n, err := escape(s[i:], &cur, &link)
		if err != nil {
			return nil, fmt.Errorf("byte %d: %w", i, err)
		}
		i += n
		start = i
	}
	if start < len(s) {
		segs = append(segs, segment{text: s[start:], attrs: cur, link: link})
	}
	return segs, nil
}

// escape applies the escape sequence at the start of s to a and link and
// returns its length.
func escape(s string, a *attrs, link *string) (int, error) {
	switch {
	case strings.HasPrefix(s, "\x1b["):
		i := 2
		for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == ';') {
			i++
		}
		if i == len(s) {
			return 0, fmt.Errorf("unterminated escape sequence %q", truncate(s))
		}
		if s[i] != 'm' {
			return 0, fmt.Errorf("unsupported escape sequence %q", s[:i+1])
		}
		if err := sgr(s[2:i], a); err != nil {
			return 0, err
		}
		return i + 1, nil
	case strings.HasPrefix(s, "\x1b]8;"):
		end := strings.Index(s, "\x1b\\")
		if end < 0 || strings.IndexByte(s[1:end], '\x1b') >= 0 {
			return 0, fmt.Errorf("unterminated hyperlink %q", truncate(s))
		}
		params, url, ok := strings.Cut(s[len("\x1b]8;"):end], ";")
		if !ok || params != "" {
			return 0, fmt.Errorf("unsupported hyperlink %q", s[:end+2])
		}
		*link = url
		return end + 2, nil
	}
	return 0, fmt.Errorf("unsupported escape sequence %q", truncate(s))
}

// sgr applies the parameters of an SGR sequence to a.
func sgr(params string, a *attrs) error {
	if params == "" {
		*a = attrs{}
		return nil
	}
	codes := strings.Split(params, ";")
	for i := 0; i < len(codes); i++ {
		n, err := strconv.Atoi(codes[i])
		if err != nil {
			return fmt.Errorf("invalid SGR parameters %q", params)
		}
		switch {
		case n == 0:
			*a = attrs{}
		case n == 1:
			a.bold = true
		case n == 2:
			a.faint = true
		case n == 3:
			a.italic = true
		case n == 4:
			a.underline = true
		case n == 9:
			a.strike = true
		case n >= 30 && n <= 37, n >= 90 && n <= 97:
			a.ansi, a.rgb = n, ""
		case n == 38 && i+4 < len(codes) && codes[i+1] == "2":
			var rgb [3]int
			for k := range rgb {
				c, err := strconv.Atoi(codes[i+2+k])
				if err != nil || c < 0 || c > 255 {
					return fmt.Errorf("invalid SGR color %q", params)
				}
				rgb[k] = c
			}
			a.ansi, a.rgb = 0, fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
			i += 4
		default:
			return fmt.Errorf("unsupported SGR parameters %q", params)
		}
	}
	return nil
}

// truncate shortens s for an error message.
func truncate(s string) string {
	const limit = 32
	if len(s) > limit {
		return s[:limit] + "..."
	}
	return s
}

// emit writes the segments as HTML: a link as <a>, a style as <span>.
func emit(segs []segment) string {
	var b strings.Builder
	link := ""
	for _, sg := range segs {
		if sg.link != link {
			if link != "" {
				b.WriteString("</a>")
			}
			if sg.link != "" {
				b.WriteString(`<a href="`)
				b.WriteString(html.EscapeString(sg.link))
				b.WriteString(`">`)
			}
			link = sg.link
		}
		open := sg.attrs.open()
		b.WriteString(open)
		writeText(&b, sg.text)
		if open != "" {
			b.WriteString("</span>")
		}
	}
	if link != "" {
		b.WriteString("</a>")
	}
	return b.String()
}

// open returns the <span> tag that starts text with these attributes, or ""
// for unstyled text.
func (a attrs) open() string {
	var classes []string
	for _, c := range []struct {
		on   bool
		name string
	}{{a.bold, "b"}, {a.faint, "f"}, {a.italic, "i"}, {a.underline, "u"}, {a.strike, "s"}} {
		if c.on {
			classes = append(classes, c.name)
		}
	}
	color := ""
	switch {
	case a.rgb != "":
		color = a.rgb
	case a.ansi != 0:
		color = "var(--ansi-" + strconv.Itoa(a.ansi) + ")"
	}
	if len(classes) == 0 && color == "" {
		return ""
	}
	tag := "<span"
	if len(classes) > 0 {
		tag += ` class="` + strings.Join(classes, " ") + `"`
	}
	if color != "" {
		tag += ` style="color:` + color + `"`
	}
	return tag + ">"
}

// writeText writes s HTML-escaped, with every grapheme cluster that is two
// columns wide in a two-cell box.
func writeText(b *strings.Builder, s string) {
	g := uniseg.NewGraphemes(s)
	for g.Next() {
		c := g.Str()
		if text.Width(c) == 2 {
			b.WriteString(`<span class="w">`)
			b.WriteString(html.EscapeString(c))
			b.WriteString(`</span>`)
			continue
		}
		b.WriteString(html.EscapeString(c))
	}
}
