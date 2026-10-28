package render

import (
	"slices"
	"strings"

	"golang.org/x/net/html"

	"github.com/ekalinin/termd/internal/style"
	"github.com/ekalinin/termd/internal/text"
)

// htmlTokens splits an HTML fragment into tokens. Text and attribute values
// are decoded, and their control characters are shown as control pictures.
// The tokenizer drops a tag that is not closed at the end of the fragment,
// such as <p align="center; that rest is returned as a text token.
func htmlTokens(s string) []html.Token {
	z := html.NewTokenizer(strings.NewReader(s))
	var toks []html.Token
	used := 0
	for z.Next() != html.ErrorToken {
		used += len(z.Raw())
		t := z.Token()
		if t.Type == html.TextToken {
			t.Data = strings.Map(controlPicture, t.Data)
		}
		for i := range t.Attr {
			t.Attr[i].Val = strings.Map(controlPicture, t.Attr[i].Val)
		}
		toks = append(toks, t)
	}
	if used < len(s) {
		toks = append(toks, html.Token{Type: html.TextToken, Data: s[used:]})
	}
	return toks
}

// htmlAttr returns the value of the attribute key of a tag token.
func htmlAttr(t html.Token, key string) (string, bool) {
	for _, a := range t.Attr {
		if a.Key == key {
			return a.Val, true
		}
	}
	return "", false
}

// blockTag reports whether the tag starts or ends a block. Inside a
// paragraph, a heading or a table cell such a tag breaks the line.
func blockTag(name string) bool {
	switch name {
	case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "details", "summary":
		return true
	}
	return false
}

// inlineHTML lays out inline content with the HTML tags between its spans.
// It keeps the tags opened in one paragraph, heading or table cell; the tags
// that are still open at its end end there.
type inlineHTML struct {
	r    *renderer
	base style.Style
	open []openTag
	out  []text.Span
	// brk is set by a block tag: the next spans start a new line, unless
	// nothing has been added yet.
	brk   bool
	added bool
}

// openTag is an open HTML tag with the style of its content. The content of
// a link is collected until the link ends; the content of script and style
// is dropped.
type openTag struct {
	name  string
	style style.Style
	skip  bool
	link  bool
	dest  string
	spans []text.Span
}

// style returns the style of the content at the current position.
func (h *inlineHTML) style() style.Style {
	if n := len(h.open); n > 0 {
		return h.open[n-1].style
	}
	return h.base
}

// add adds the spans of content at the current position.
func (h *inlineHTML) add(spans ...text.Span) {
	if len(spans) == 0 {
		return
	}
	for _, t := range h.open {
		if t.skip {
			return
		}
	}
	if h.brk && h.added {
		h.put(text.Break)
	}
	h.brk, h.added = false, true
	h.put(spans...)
}

// put appends spans to the innermost open link, or to the output.
func (h *inlineHTML) put(spans ...text.Span) {
	for i, t := range slices.Backward(h.open) {
		if t.link {
			h.open[i].spans = append(h.open[i].spans, spans...)
			return
		}
	}
	h.out = append(h.out, spans...)
}

// token applies an HTML token at the current position.
func (h *inlineHTML) token(t html.Token) {
	switch t.Type {
	case html.TextToken:
		// A line break of HTML text is whitespace; a span of just "\n"
		// would be a forced break.
		h.add(text.Span{Text: strings.ReplaceAll(t.Data, "\n", " "), Style: h.style()})
	case html.StartTagToken, html.SelfClosingTagToken:
		h.start(t)
	case html.EndTagToken:
		h.end(t.Data)
	}
}

func (h *inlineHTML) start(t html.Token) {
	st := h.style()
	switch t.Data {
	case "img":
		alt, _ := htmlAttr(t, "alt")
		src, _ := htmlAttr(t, "src")
		h.add(h.r.image(alt, h.r.htmlDestination(src), st)...)
	case "br":
		h.add(text.Break)
	case "b", "strong":
		st.Bold = true
		h.open = append(h.open, openTag{name: t.Data, style: st})
	case "i", "em":
		st.Italic = true
		h.open = append(h.open, openTag{name: t.Data, style: st})
	case "kbd", "code":
		h.open = append(h.open, openTag{name: t.Data, style: style.Layer(st, h.r.pal.InlineCode)})
	case "a":
		tag := openTag{name: t.Data, style: st}
		if href, ok := htmlAttr(t, "href"); ok {
			tag.style, tag.link, tag.dest = style.Layer(st, h.r.pal.Link), true, h.r.htmlDestination(href)
		}
		h.open = append(h.open, tag)
	case "script", "style":
		h.open = append(h.open, openTag{name: t.Data, style: st, skip: true})
	default:
		h.brk = h.brk || blockTag(t.Data)
	}
}

// end closes the innermost open tag named name and the tags opened after it.
// An end tag without an open tag is ignored.
func (h *inlineHTML) end(name string) {
	if blockTag(name) {
		h.brk = true
		return
	}
	for i, t := range slices.Backward(h.open) {
		if t.name == name {
			for len(h.open) > i {
				h.pop()
			}
			return
		}
	}
}

// pop closes the innermost open tag; a link passes its content on as a
// link to its destination.
func (h *inlineHTML) pop() {
	t := h.open[len(h.open)-1]
	h.open = h.open[:len(h.open)-1]
	if t.link {
		h.put(h.r.link(t.spans, t.dest, false)...)
	}
}

// spans closes the open tags and returns the laid out content.
func (h *inlineHTML) spans() []text.Span {
	for len(h.open) > 0 {
		h.pop()
	}
	return h.out
}

// headingLevel returns the level of an h1 to h6 tag, 0 for any other tag.
func headingLevel(name string) int {
	if len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6' {
		return int(name[1] - '0')
	}
	return 0
}

// htmlBlocks lays out the blocks of a container with the HTML blocks among
// them: the paragraphs and headings of an HTML block become blocks of their
// own, empty ones are dropped, and a details holds the blocks between its
// start and end tags.
type htmlBlocks struct {
	r     *renderer
	width int
	// frames holds the blocks of the container and of every open details
	// inside it, outermost first.
	frames []*detailsFrame
	// part is the paragraph, heading or summary being filled, nil between
	// them; level is the level of a heading, 0 otherwise.
	part    *inlineHTML
	level   int
	summary bool
}

// detailsFrame is the content of a container or of an open details.
type detailsFrame struct {
	blocks     []Block
	summary    []text.Span
	hasSummary bool
}

func newHTMLBlocks(r *renderer, width int) *htmlBlocks {
	return &htmlBlocks{r: r, width: width, frames: []*detailsFrame{{}}}
}

// contentWidth is the width of a block at the current position: every open
// details indents its content by 2 columns.
func (b *htmlBlocks) contentWidth() int {
	return b.width - 2*(len(b.frames)-1)
}

// add adds a block at the current position.
func (b *htmlBlocks) add(blk Block) {
	f := b.frames[len(b.frames)-1]
	f.blocks = append(f.blocks, blk)
}

// html lays out the source of an HTML block. Tags that are not closed in it
// end with it, except details.
func (b *htmlBlocks) html(src string) {
	for _, t := range htmlTokens(src) {
		b.token(t)
	}
	b.flush()
	b.level, b.summary = 0, false
}

func (b *htmlBlocks) token(t html.Token) {
	if t.Type == html.TextToken || t.Type == html.CommentToken || !blockTag(t.Data) {
		if b.part == nil {
			base := style.Style{}
			if b.level > 0 {
				base = b.r.pal.Heading
			}
			b.part = &inlineHTML{r: b.r, base: base}
		}
		b.part.token(t)
		return
	}
	b.flush()
	b.level, b.summary = 0, false
	start := t.Type != html.EndTagToken
	switch {
	case t.Data == "details" && start:
		b.frames = append(b.frames, &detailsFrame{})
	case t.Data == "details":
		if len(b.frames) > 1 {
			b.closeDetails()
		}
	case t.Data == "summary" && start:
		// Only the first summary of a details is its summary; any other
		// one is a paragraph.
		f := b.frames[len(b.frames)-1]
		b.summary = len(b.frames) > 1 && !f.hasSummary
	case start:
		b.level = headingLevel(t.Data)
	}
}

// flush ends the paragraph, heading or summary being filled.
func (b *htmlBlocks) flush() {
	if b.part == nil {
		return
	}
	spans := b.part.spans()
	b.part = nil
	if b.summary {
		f := b.frames[len(b.frames)-1]
		f.summary, f.hasSummary = spans, true
		return
	}
	if text.NaturalWidth(spans) == 0 {
		return
	}
	if b.level > 0 {
		spans = append([]text.Span{{Text: strings.Repeat("#", b.level) + " ", Style: b.r.pal.Heading}}, spans...)
	}
	b.add(Block{Lines: text.Wrap(spans, b.contentWidth(), true)})
}

// closeDetails closes the innermost open details and adds it to the content
// around it: the summary after the marker, then the content indented by 2
// columns. A details without a summary shows Details.
func (b *htmlBlocks) closeDetails() {
	f := b.frames[len(b.frames)-1]
	w := b.contentWidth()
	b.frames = b.frames[:len(b.frames)-1]

	summary := f.summary
	if text.NaturalWidth(summary) == 0 {
		summary = []text.Span{{Text: "Details"}}
	}
	blk := Block{Wide: anyWide(f.blocks)}
	for i, l := range text.Wrap(summary, w, true) {
		switch {
		case i == 0:
			l = l.Prepend(text.Span{Text: "▾ ", Style: b.r.pal.Marker})
		case len(l) > 0:
			l = l.Prepend(text.Span{Text: "  "})
		}
		blk.Lines = append(blk.Lines, l)
	}
	for _, l := range join(f.blocks, true) {
		if len(l) > 0 {
			l = l.Prepend(text.Span{Text: "  "})
		}
		blk.Lines = append(blk.Lines, l)
	}
	b.add(blk)
}

// end closes the details that are still open and returns the blocks of the
// container.
func (b *htmlBlocks) end() []Block {
	b.flush()
	for len(b.frames) > 1 {
		b.closeDetails()
	}
	return b.frames[0].blocks
}
