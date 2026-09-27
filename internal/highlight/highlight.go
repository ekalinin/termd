package highlight

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
)

// Theme selects the color set used for highlighting.
type Theme int

const (
	// Dark suits terminals with a dark background.
	Dark Theme = iota
	// Light suits terminals with a light background.
	Light
)

// styleName returns the chroma style used for the theme.
func (t Theme) styleName() string {
	if t == Light {
		return "github"
	}
	return "github-dark"
}

// Language returns the lower-cased first word of a code block info string.
func Language(info string) string {
	fields := strings.Fields(info)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0])
}

// lexer returns the chroma lexer named by the info string, or nil. Lookup
// uses lexer names, aliases and file extensions only; content is never
// analysed, because guessing produces wrong colors.
func lexer(info string) chroma.Lexer {
	lang := Language(info)
	if lang == "" {
		return nil
	}
	return lexers.Get(lang)
}

// Recognized reports whether code with this info string can be highlighted.
func Recognized(info string) bool {
	return lexer(info) != nil
}

// Highlight splits code into lines of colored spans. It returns nil when the
// language is not recognized or tokenizing fails. The text of the returned
// lines is exactly the text of code.
func Highlight(code, info string, theme Theme) []text.Line {
	lx := lexer(info)
	if lx == nil {
		return nil
	}
	it, err := chroma.Coalesce(lx).Tokenise(nil, code)
	if err != nil {
		return nil
	}
	st := styles.Get(theme.styleName())
	base := st.Get(chroma.Background).Colour

	lines := []text.Line{{}}
	for tok := it(); tok != chroma.EOF; tok = it() {
		s := tokenStyle(st.Get(tok.Type), base)
		for i, part := range strings.Split(tok.Value, "\n") {
			if i > 0 {
				lines = append(lines, text.Line{})
			}
			if part != "" {
				last := len(lines) - 1
				lines[last] = append(lines[last], text.Span{Text: part, Style: s})
			}
		}
	}
	// Lexers may append a final newline; keep exactly the lines of code.
	if want := strings.Count(code, "\n") + 1; len(lines) > want {
		lines = lines[:want]
	}
	return lines
}

// tokenStyle converts a chroma style entry. Tokens drawn in the style's
// default text color keep the terminal's own foreground.
func tokenStyle(e chroma.StyleEntry, base chroma.Colour) style.Style {
	s := style.Style{
		Bold:      e.Bold == chroma.Yes,
		Italic:    e.Italic == chroma.Yes,
		Underline: e.Underline == chroma.Yes,
	}
	if e.Colour.IsSet() && e.Colour != base {
		s.FG = style.RGB(e.Colour.Red(), e.Colour.Green(), e.Colour.Blue())
	}
	return s
}
