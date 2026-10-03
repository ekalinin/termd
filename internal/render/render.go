package render

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	gtext "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/ekalinin/termd/internal/diagram"
	"github.com/ekalinin/termd/internal/highlight"
	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/table"
	"github.com/ekalinin/termd/internal/text"
	"github.com/ekalinin/termd/internal/theme"
)

// Options controls the layout and the escape sequences of the output.
type Options struct {
	// Width is the output width in columns.
	Width int
	Style style.Options
	// Theme returns the highlighting theme. It is called at most once, and
	// only when a code block is about to be highlighted.
	Theme func() highlight.Theme
	// Palette styles the elements of the document outside code blocks. The
	// zero value means theme.Default.
	Palette theme.Palette
	// Dir is the absolute directory of the document file. With hyperlinks,
	// relative link destinations are resolved against it; empty keeps every
	// destination as written.
	Dir string
}

// Block is a rendered block: its lines and whether it is wider than the
// output (a wide block is never wrapped or truncated).
type Block struct {
	Lines []text.Line
	Wide  bool
}

// Parse parses src as CommonMark with the GFM extensions and footnotes.
func Parse(src []byte) ast.Node {
	md := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote))
	return md.Parser().Parse(gtext.NewReader(src))
}

// Render lays out a markdown document and returns the terminal output. A
// leading YAML frontmatter block is shown before the document. Control
// characters of the document are shown as visible characters.
func Render(src []byte, opts Options) string {
	fm, body := splitFrontmatter(cleanSource(src))
	r := &renderer{src: body, opts: opts, pal: opts.palette()}
	var blocks []Block
	if b, ok := frontmatterBlock(fm, opts.Width, r.pal); ok {
		blocks = append(blocks, b)
	}
	lines := join(append(blocks, r.blocks(Parse(body), opts.Width)...), true)
	return output(lines, opts.Style)
}

// Code lays out the whole text of a source file as one code block and returns
// the terminal output. The file name selects the highlighting language. The
// text is never parsed as markdown and never drawn as a diagram. Control
// characters are cleaned as in a markdown document.
func Code(src []byte, name string, opts Options) string {
	if len(src) == 0 {
		return ""
	}
	code := strings.TrimSuffix(string(cleanSource(src)), "\n")
	r := &renderer{opts: opts}
	var lines []text.Line
	if opts.Style.Styled && highlight.RecognizedFile(name) {
		lines = highlight.HighlightFile(code, name, r.resolveTheme())
	}
	if lines == nil {
		lines = verbatim(code, opts.Width).Lines
	}
	return output(lines, opts.Style)
}

// output renders lines with their escape sequences, each followed by a line
// break.
func output(lines []text.Line, o style.Options) string {
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(l.Render(o))
		b.WriteByte('\n')
	}
	return b.String()
}

type renderer struct {
	src      []byte
	opts     Options
	pal      theme.Palette
	theme    highlight.Theme
	themeSet bool
}

// palette returns the palette of the options, theme.Default for a zero one.
func (o Options) palette() theme.Palette {
	if o.Palette == (theme.Palette{}) {
		return theme.Default
	}
	return o.Palette
}

// resolveTheme asks for the theme once, when first needed.
func (r *renderer) resolveTheme() highlight.Theme {
	if !r.themeSet {
		r.theme = highlight.Dark
		if r.opts.Theme != nil {
			r.theme = r.opts.Theme()
		}
		r.themeSet = true
	}
	return r.theme
}

// join concatenates blocks, separating them with a blank line when blank is
// set.
func join(blocks []Block, blank bool) []text.Line {
	var lines []text.Line
	for i, b := range blocks {
		if i > 0 && blank {
			lines = append(lines, text.Line{})
		}
		lines = append(lines, b.Lines...)
	}
	return lines
}

// blocks renders the block children of parent.
func (r *renderer) blocks(parent ast.Node, width int) []Block {
	var out []Block
	for n := parent.FirstChild(); n != nil; n = n.NextSibling() {
		if b, ok := r.block(n, width); ok {
			out = append(out, b)
		}
	}
	return out
}

func (r *renderer) block(n ast.Node, width int) (Block, bool) {
	switch n := n.(type) {
	case *ast.Heading:
		hs := r.pal.Heading
		spans := append([]text.Span{{Text: strings.Repeat("#", n.Level) + " ", Style: hs}}, r.inlines(n, hs)...)
		return Block{Lines: text.Wrap(spans, width, true)}, true
	case *ast.Paragraph, *ast.TextBlock:
		return Block{Lines: text.Wrap(r.inlines(n, style.Style{}), width, true)}, true
	case *ast.ThematicBreak:
		return rule(width, r.pal.Marker), true
	case *ast.FencedCodeBlock:
		info := ""
		if n.Info != nil {
			info = string(n.Info.Segment.Value(r.src))
		}
		return r.code(r.rawLines(n), info, width), true
	case *ast.CodeBlock:
		return r.code(r.rawLines(n), "", width), true
	case *ast.HTMLBlock:
		src := r.rawLines(n)
		if n.HasClosure() {
			src += string(n.ClosureLine.Value(r.src))
		}
		return verbatim(strings.TrimRight(src, "\n"), width), true
	case *ast.List:
		return r.list(n, width), true
	case *ast.Blockquote:
		return r.quote(n, width), true
	case *east.Table:
		return r.table(n, width), true
	case *east.FootnoteList:
		return r.footnotes(n, width), true
	}
	if n.HasChildren() {
		blocks := r.blocks(n, width)
		return Block{Lines: join(blocks, true), Wide: anyWide(blocks)}, true
	}
	return Block{}, false
}

// rule is the horizontal line of a thematic break in the marker style ms.
func rule(width int, ms style.Style) Block {
	return Block{Lines: []text.Line{{{Text: strings.Repeat("─", width), Style: ms}}}}
}

func anyWide(blocks []Block) bool {
	for _, b := range blocks {
		if b.Wide {
			return true
		}
	}
	return false
}

// rawLines returns the source lines of a block node.
func (r *renderer) rawLines(n ast.Node) string {
	var b strings.Builder
	lines := n.Lines()
	for i := range lines.Len() {
		seg := lines.At(i)
		b.Write(seg.Value(r.src))
	}
	return b.String()
}

// verbatim turns text into unwrapped plain lines.
func verbatim(s string, width int) Block {
	var b Block
	for l := range strings.SplitSeq(s, "\n") {
		line := text.Plain(l)
		b.Lines = append(b.Lines, line)
		b.Wide = b.Wide || line.Width() > width
	}
	return b
}

// code renders a code block: diagrams go to the diagram renderer, other
// code is shown verbatim and highlighted when styling allows.
func (r *renderer) code(src, info string, width int) Block {
	code := strings.TrimSuffix(src, "\n")
	if lang, ok := diagram.Language(info); ok {
		res := diagram.Render(lang, code, width)
		return Block{Lines: res.Lines, Wide: res.Wide}
	}
	if r.opts.Style.Styled && highlight.Recognized(info) {
		if lines := highlight.Highlight(code, info, r.resolveTheme()); lines != nil {
			b := Block{Lines: lines}
			for _, l := range lines {
				b.Wide = b.Wide || l.Width() > width
			}
			return b
		}
	}
	return verbatim(code, width)
}

// list renders list items with a bullet or number; continuation lines are
// indented to the start of the item text.
func (r *renderer) list(n *ast.List, width int) Block {
	var markers []string
	num := n.Start
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		if n.IsOrdered() {
			markers = append(markers, strconv.Itoa(num)+string(n.Marker))
			num++
		} else {
			markers = append(markers, "•")
		}
	}
	return r.items(n, markers, n.IsTight, width)
}

// items lays out the children of parent as list items with the markers
// aligned to the right. A loose list separates the items and the blocks
// inside them with a blank line.
func (r *renderer) items(parent ast.Node, markers []string, tight bool, width int) Block {
	markerWidth := 0
	for _, m := range markers {
		markerWidth = max(markerWidth, text.Width(m))
	}
	indent := markerWidth + 1

	var b Block
	i := 0
	for item := parent.FirstChild(); item != nil; item = item.NextSibling() {
		if i > 0 && !tight {
			b.Lines = append(b.Lines, text.Line{})
		}
		children := r.blocks(item, width-indent)
		b.Wide = b.Wide || anyWide(children)
		lines := join(children, !tight)
		if len(lines) == 0 {
			lines = []text.Line{{}}
		}
		marker := strings.Repeat(" ", markerWidth-text.Width(markers[i])) + markers[i] + " "
		for j, l := range lines {
			switch {
			case j == 0:
				l = l.Prepend(text.Span{Text: marker, Style: r.pal.Marker})
			case len(l) > 0:
				l = l.Prepend(text.Span{Text: strings.Repeat(" ", indent)})
			}
			b.Lines = append(b.Lines, l)
		}
		i++
	}
	return b
}

// alertType is a GitHub alert: the marker that starts the quote, the title
// shown in its place and the palette style of the title and the quote
// marker.
type alertType struct {
	marker, title string
	color         func(theme.Palette) style.Style
}

var alertTypes = []alertType{
	{"[!NOTE]", "Note", func(p theme.Palette) style.Style { return p.Note }},
	{"[!TIP]", "Tip", func(p theme.Palette) style.Style { return p.Tip }},
	{"[!IMPORTANT]", "Important", func(p theme.Palette) style.Style { return p.Important }},
	{"[!WARNING]", "Warning", func(p theme.Palette) style.Style { return p.Warning }},
	{"[!CAUTION]", "Caution", func(p theme.Palette) style.Style { return p.Caution }},
}

// alert reports whether the block quote n is a GitHub alert: its first child
// is a paragraph whose first line is an alert marker alone, in any case. It
// returns the alert type and the spans of the rest of the paragraph.
func (r *renderer) alert(n *ast.Blockquote) (alertType, []text.Span, bool) {
	p, ok := n.FirstChild().(*ast.Paragraph)
	if !ok {
		return alertType{}, nil, false
	}
	// The marker is matched in the raw source, so an escaped \[!NOTE] stays
	// a regular quote.
	var line []byte
	c := p.FirstChild()
	for c != nil {
		t, ok := c.(*ast.Text)
		if !ok {
			return alertType{}, nil, false
		}
		line = append(line, t.Segment.Value(r.src)...)
		c = c.NextSibling()
		if t.SoftLineBreak() || t.HardLineBreak() {
			break
		}
	}
	marker := strings.TrimSpace(string(line))
	for _, a := range alertTypes {
		if strings.EqualFold(marker, a.marker) {
			var rest []text.Span
			for ; c != nil; c = c.NextSibling() {
				rest = append(rest, r.inline(c, style.Style{})...)
			}
			return a, rest, true
		}
	}
	return alertType{}, nil, false
}

// quote prefixes every line of a block quote with a quote marker. In a GitHub
// alert the title takes the place of the marker line, and the title and the
// quote marker have the color of the alert.
func (r *renderer) quote(n *ast.Blockquote, width int) Block {
	children := r.blocks(n, width-2)
	ms := r.pal.Marker
	if a, rest, ok := r.alert(n); ok {
		ms = a.color(r.pal)
		spans := []text.Span{{Text: a.title, Style: style.Layer(ms, style.Style{Bold: true})}}
		if len(rest) > 0 {
			spans = append(append(spans, text.Break), rest...)
		}
		children[0] = Block{Lines: text.Wrap(spans, width-2, true)}
	}
	b := Block{Wide: anyWide(children)}
	for _, l := range join(children, true) {
		if len(l) == 0 {
			b.Lines = append(b.Lines, text.Line{{Text: "│", Style: ms}})
			continue
		}
		b.Lines = append(b.Lines, l.Prepend(text.Span{Text: "│ ", Style: ms}))
	}
	return b
}

// table converts a GFM table node and lays it out.
func (r *renderer) table(n *east.Table, width int) Block {
	t := table.Table{HeaderStyle: r.pal.TableHeader, BorderStyle: r.pal.TableBorder}
	for _, a := range n.Alignments {
		switch a {
		case east.AlignLeft:
			t.Align = append(t.Align, table.AlignLeft)
		case east.AlignCenter:
			t.Align = append(t.Align, table.AlignCenter)
		case east.AlignRight:
			t.Align = append(t.Align, table.AlignRight)
		default:
			t.Align = append(t.Align, table.AlignNone)
		}
	}
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		var cells []table.Cell
		for c := row.FirstChild(); c != nil; c = c.NextSibling() {
			cells = append(cells, table.Cell(r.inlines(c, style.Style{})))
		}
		if _, ok := row.(*east.TableHeader); ok {
			t.Header = cells
		} else {
			t.Rows = append(t.Rows, cells)
		}
	}
	lines, wide := t.Render(width)
	return Block{Lines: lines, Wide: wide}
}

// footnotes renders the footnote definitions after a horizontal line as a
// numbered list. goldmark keeps only the referenced definitions, ordered by
// their numbers. Like a markdown list, the list is loose when a definition
// has more than one block.
func (r *renderer) footnotes(n *east.FootnoteList, width int) Block {
	var markers []string
	tight := true
	for fn := n.FirstChild(); fn != nil; fn = fn.NextSibling() {
		markers = append(markers, strconv.Itoa(fn.(*east.Footnote).Index)+".")
		blocks := 0
		for c := fn.FirstChild(); c != nil; c = c.NextSibling() {
			if _, ok := c.(*east.FootnoteBacklink); !ok {
				blocks++
			}
		}
		tight = tight && blocks <= 1
	}
	list := r.items(n, markers, tight, width)
	return Block{Lines: join([]Block{rule(width, r.pal.Marker), list}, true), Wide: list.Wide}
}

// inlines renders the inline children of n on top of the base style st.
func (r *renderer) inlines(n ast.Node, st style.Style) []text.Span {
	var out []text.Span
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		out = append(out, r.inline(c, st)...)
	}
	return out
}

func (r *renderer) inline(n ast.Node, st style.Style) []text.Span {
	hyperlinks := r.opts.Style.Hyperlinks
	switch n := n.(type) {
	case *ast.Text:
		v := n.Segment.Value(r.src)
		if !n.IsRaw() {
			v = unescape(v)
		}
		spans := []text.Span{{Text: string(v), Style: st}}
		switch {
		case n.HardLineBreak():
			spans = append(spans, text.Break)
		case n.SoftLineBreak():
			spans = append(spans, text.Span{Text: " ", Style: st})
		}
		return spans
	case *ast.String:
		return []text.Span{{Text: string(n.Value), Style: st}}
	case *ast.CodeSpan:
		cs := style.Layer(st, r.pal.InlineCode)
		return []text.Span{{Text: r.plainText(n, true), Style: cs}}
	case *ast.Emphasis:
		es := st
		if n.Level >= 2 {
			es.Bold = true
		} else {
			es.Italic = true
		}
		return r.inlines(n, es)
	case *east.Strikethrough:
		ss := st
		ss.Strike = true
		return r.inlines(n, ss)
	case *ast.Link:
		return text.LinkSpans(r.inlines(n, style.Layer(st, r.pal.Link)), r.destination(n.Destination), hyperlinks)
	case *ast.AutoLink:
		label := string(n.Label(r.src))
		url := string(n.URL(r.src))
		if n.AutoLinkType == ast.AutoLinkEmail {
			if !hyperlinks {
				url = label
			} else if !strings.HasPrefix(strings.ToLower(url), "mailto:") {
				url = "mailto:" + url
			}
		}
		return text.LinkSpans([]text.Span{{Text: label, Style: style.Layer(st, r.pal.Link)}}, url, hyperlinks)
	case *ast.Image:
		label := []text.Span{{Text: "[image: " + r.plainText(n, false) + "]", Style: style.Layer(st, r.pal.Link)}}
		return text.LinkSpans(label, r.destination(n.Destination), hyperlinks)
	case *ast.RawHTML:
		var b strings.Builder
		for i := range n.Segments.Len() {
			seg := n.Segments.At(i)
			b.Write(seg.Value(r.src))
		}
		return []text.Span{{Text: b.String(), Style: st}}
	case *east.TaskCheckBox:
		if n.IsChecked {
			return []text.Span{{Text: "[x] ", Style: st}}
		}
		return []text.Span{{Text: "[ ] ", Style: st}}
	case *east.FootnoteLink:
		// Not a hyperlink: a terminal link cannot jump to the definition.
		return []text.Span{{Text: "[" + strconv.Itoa(n.Index) + "]", Style: st}}
	case *east.FootnoteBacklink:
		// Back-references are not shown.
		return nil
	}
	return r.inlines(n, st)
}

// plainText collects the text of n's descendants. Raw text (code spans) is
// kept as written; line breaks become spaces.
func (r *renderer) plainText(n ast.Node, raw bool) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *ast.Text:
			v := c.Segment.Value(r.src)
			if !raw && !c.IsRaw() {
				v = unescape(v)
			}
			b.Write(v)
			if c.SoftLineBreak() || c.HardLineBreak() {
				b.WriteByte(' ')
			}
		case *ast.String:
			b.Write(c.Value)
		}
		return ast.WalkContinue, nil
	})
	return strings.ReplaceAll(b.String(), "\n", " ")
}

// unescape resolves backslash escapes and character references the way the
// HTML renderer of goldmark does. A reference to a control character, for
// example &#27;, gives its control picture, as in the source.
func unescape(v []byte) []byte {
	v = util.UnescapePunctuations(v)
	v = util.ResolveNumericReferences(v)
	return bytes.Map(controlPicture, util.ResolveEntityNames(v))
}
