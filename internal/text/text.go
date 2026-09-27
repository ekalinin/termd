package text

import (
	"strings"
	"unicode"

	"github.com/rivo/uniseg"

	"github.com/ekalinin/termd/internal/style"
)

// Width returns the display width of s in terminal columns, measured by
// grapheme cluster. s must not contain escape sequences.
func Width(s string) int {
	return uniseg.StringWidth(s)
}

// Span is a run of text with one style and an optional link destination.
type Span struct {
	Text  string
	Style style.Style
	Link  string
}

// Break is a span that forces a line break when wrapping.
var Break = Span{Text: "\n"}

// Line is a sequence of spans shown on one terminal line.
type Line []Span

// Plain returns a line holding s without style.
func Plain(s string) Line {
	if s == "" {
		return Line{}
	}
	return Line{{Text: s}}
}

// Width returns the display width of the line; escape sequences never
// contribute because only span text is measured.
func (l Line) Width() int {
	w := 0
	for _, sp := range l {
		w += Width(sp.Text)
	}
	return w
}

// String returns the text of the line without any escape sequences.
func (l Line) String() string {
	var b strings.Builder
	for _, sp := range l {
		b.WriteString(sp.Text)
	}
	return b.String()
}

// Render returns the line with the escape sequences enabled by o. Adjacent
// spans with the same style and link share one escape sequence.
func (l Line) Render(o style.Options) string {
	var b, run strings.Builder
	var runStyle style.Style
	link := ""
	flush := func() {
		if run.Len() > 0 {
			b.WriteString(o.Format(run.String(), runStyle))
			run.Reset()
		}
	}
	for _, sp := range l {
		if sp.Link != link {
			flush()
			if link != "" {
				b.WriteString(o.LinkClose())
			}
			if sp.Link != "" {
				b.WriteString(o.LinkOpen(sp.Link))
			}
			link = sp.Link
		}
		if sp.Style != runStyle {
			flush()
			runStyle = sp.Style
		}
		run.WriteString(sp.Text)
	}
	flush()
	if link != "" {
		b.WriteString(o.LinkClose())
	}
	return b.String()
}

// Prepend returns a new line starting with prefix spans followed by l.
func (l Line) Prepend(prefix ...Span) Line {
	out := make(Line, 0, len(prefix)+len(l))
	out = append(out, prefix...)
	return append(out, l...)
}

// word is a run of non-space text that may span several styles.
type word struct {
	pieces []Span
	width  int
}

// token is a word, a whitespace run or a forced break.
type token struct {
	kind  int
	word  word
	space Span
}

const (
	tokWord = iota
	tokSpace
	tokBreak
)

// tokenize splits spans into words, whitespace runs and forced breaks.
// Whitespace inside a span keeps that span's style and link, so a space in
// the middle of a link stays part of the link.
func tokenize(spans []Span) []token {
	var toks []token
	var cur word
	var piece strings.Builder
	var pieceSpan Span

	flushPiece := func() {
		if piece.Len() == 0 {
			return
		}
		sp := pieceSpan
		sp.Text = piece.String()
		cur.pieces = append(cur.pieces, sp)
		cur.width += Width(sp.Text)
		piece.Reset()
	}
	flushWord := func() {
		flushPiece()
		if len(cur.pieces) > 0 {
			toks = append(toks, token{kind: tokWord, word: cur})
		}
		cur = word{}
	}

	for _, sp := range spans {
		if sp.Text == "\n" {
			flushWord()
			toks = append(toks, token{kind: tokBreak})
			continue
		}
		pieceSpan = sp
		for _, r := range sp.Text {
			switch {
			case r == '\n':
				flushWord()
				toks = append(toks, token{kind: tokSpace, space: Span{Text: " ", Style: sp.Style, Link: sp.Link}})
			case unicode.IsSpace(r):
				flushWord()
				if n := len(toks); n == 0 || toks[n-1].kind != tokSpace {
					toks = append(toks, token{kind: tokSpace, space: Span{Text: " ", Style: sp.Style, Link: sp.Link}})
				}
			default:
				piece.WriteRune(r)
			}
		}
		flushPiece()
	}
	flushWord()
	return toks
}

// Wrap lays out spans into lines no wider than width, breaking at
// whitespace. Runs of whitespace collapse to a single space and Break spans
// force a new line. When breakLong is true, a word wider than width is
// split at grapheme boundaries; otherwise it is kept whole and overflows.
func Wrap(spans []Span, width int, breakLong bool) []Line {
	if width < 1 {
		width = 1
	}
	var lines []Line
	var cur Line
	curW := 0
	var pending *Span

	flush := func() {
		lines = append(lines, cur)
		cur = nil
		curW = 0
	}

	for _, tok := range tokenize(spans) {
		switch tok.kind {
		case tokBreak:
			flush()
			pending = nil
		case tokSpace:
			if curW > 0 && pending == nil {
				sp := tok.space
				pending = &sp
			}
		case tokWord:
			w := tok.word
			gap := 0
			if pending != nil && curW > 0 {
				gap = 1
			}
			switch {
			case curW+gap+w.width <= width:
				if gap == 1 {
					cur = append(cur, *pending)
				}
				cur = append(cur, w.pieces...)
				curW += gap + w.width
			case w.width <= width || !breakLong:
				if curW > 0 {
					flush()
				}
				cur = append(cur, w.pieces...)
				curW = w.width
			default:
				if curW > 0 {
					flush()
				}
				for _, chunk := range splitWord(w, width) {
					if curW > 0 {
						flush()
					}
					cur = append(cur, chunk.pieces...)
					curW = chunk.width
				}
			}
			pending = nil
		}
	}
	if len(cur) > 0 || len(lines) == 0 {
		lines = append(lines, cur)
	}
	return lines
}

// splitWord cuts a word into chunks no wider than width at grapheme
// cluster boundaries.
func splitWord(w word, width int) []word {
	var chunks []word
	var cur word
	for _, sp := range w.pieces {
		var b strings.Builder
		g := uniseg.NewGraphemes(sp.Text)
		for g.Next() {
			cw := g.Width()
			if cur.width+cw > width && (cur.width > 0 || b.Len() > 0) {
				if b.Len() > 0 {
					p := sp
					p.Text = b.String()
					cur.pieces = append(cur.pieces, p)
					b.Reset()
				}
				chunks = append(chunks, cur)
				cur = word{}
			}
			b.WriteString(g.Str())
			cur.width += cw
		}
		if b.Len() > 0 {
			p := sp
			p.Text = b.String()
			cur.pieces = append(cur.pieces, p)
		}
	}
	if len(cur.pieces) > 0 {
		chunks = append(chunks, cur)
	}
	return chunks
}

// MaxWordWidth returns the display width of the widest word in spans.
func MaxWordWidth(spans []Span) int {
	widest := 0
	for _, tok := range tokenize(spans) {
		if tok.kind == tokWord {
			widest = max(widest, tok.word.width)
		}
	}
	return widest
}

// NaturalWidth returns the width of spans laid out on one line with
// whitespace collapsed. Forced breaks start a new line; the widest line wins.
func NaturalWidth(spans []Span) int {
	widest := 0
	for _, l := range Wrap(spans, 1<<30, false) {
		widest = max(widest, l.Width())
	}
	return widest
}

// linkStyle marks link text when styling is enabled.
var linkStyle = style.Style{Underline: true, ANSI: 34}

// LinkSpans turns the label of a link into spans. With hyperlinks the label
// carries the destination as an OSC 8 link; without them the destination is
// appended as " (url)", unless the label already reads as the URL.
func LinkSpans(label []Span, url string, hyperlinks bool) []Span {
	out := make([]Span, 0, len(label)+1)
	for _, sp := range label {
		if sp.Style.ANSI == 0 && !sp.Style.FG.Set {
			sp.Style.ANSI = linkStyle.ANSI
		}
		sp.Style.Underline = true
		if hyperlinks {
			sp.Link = url
		}
		out = append(out, sp)
	}
	if hyperlinks || url == "" || Line(label).String() == url {
		return out
	}
	return append(out, Span{Text: " (" + url + ")"})
}

// Pad returns l extended with spaces to width according to align:
// 'l' left, 'r' right, 'c' center. Lines wider than width are unchanged.
func Pad(l Line, width int, align byte) Line {
	gap := width - l.Width()
	if gap <= 0 {
		return l
	}
	left, right := 0, gap
	switch align {
	case 'r':
		left, right = gap, 0
	case 'c':
		left = gap / 2
		right = gap - left
	}
	out := make(Line, 0, len(l)+2)
	if left > 0 {
		out = append(out, Span{Text: strings.Repeat(" ", left)})
	}
	out = append(out, l...)
	if right > 0 {
		out = append(out, Span{Text: strings.Repeat(" ", right)})
	}
	return out
}
